package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/ungerr"
)

type sessionService struct {
	jwtService       auth.JWTService
	users            auth.UserStore
	transactor       auth.Transactor
	sessions         auth.SessionStore
	refreshTokens    auth.RefreshTokenStore
	refreshTokenTTL  time.Duration
	claimsBuilder    func(ctx context.Context, userID string, baseClaims map[string]any) (map[string]any, error)
}

func NewSessionService(
	jwtService auth.JWTService,
	users auth.UserStore,
	transactor auth.Transactor,
	sessions auth.SessionStore,
	refreshTokens auth.RefreshTokenStore,
	refreshTokenTTL time.Duration,
	claimsBuilder func(ctx context.Context, userID string, baseClaims map[string]any) (map[string]any, error),
) SessionService {
	return &sessionService{
		jwtService,
		users,
		transactor,
		sessions,
		refreshTokens,
		refreshTokenTTL,
		claimsBuilder,
	}
}

// RefreshToken validates and rotates a refresh token, issuing new access and refresh tokens
func (ss *sessionService) RefreshToken(ctx context.Context, request dto.RefreshTokenRequest) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.RefreshToken")
	defer span.End()

	var response dto.TokenResponse

	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		refreshToken, err := ss.getRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			return err
		}

		session, err := ss.GetByID(ctx, refreshToken.SessionID)
		if err != nil {
			return err
		}

		// Verify user
		if err := ss.users.Exists(ctx, session.UserID); err != nil {
			return err
		}

		// Rotate the refresh token (this validates expiry and deletes old token)
		newRefreshToken, err := ss.rotateRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			return err
		}

		rawFingerprint, fgpHash := ss.generateFingerprint()
		claims := mapper.SessionToAuthData(session, fgpHash)

		if ss.claimsBuilder != nil {
			builtClaims, err := ss.claimsBuilder(ctx, session.UserID, claims)
			if err != nil {
				return err
			}
			claims = builtClaims
		}

		accessToken, err := ss.jwtService.CreateToken(claims)
		if err != nil {
			return err
		}

		response = dto.NewTokenResp(accessToken, newRefreshToken, rawFingerprint)
		return nil
	})

	return response, err
}

func (ss *sessionService) CreateTokenAndSession(ctx context.Context, user auth.User) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.CreateTokenAndSession")
	defer span.End()

	// Create session with refresh token
	session, refreshToken, err := ss.createSession(ctx, user.ID, ss.refreshTokenTTL)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	rawFingerprint, fgpHash := ss.generateFingerprint()
	authData := mapper.SessionToAuthData(session, fgpHash)

	if ss.claimsBuilder != nil {
		builtClaims, err := ss.claimsBuilder(ctx, session.UserID, authData)
		if err != nil {
			return dto.TokenResponse{}, err
		}
		authData = builtClaims
	}

	accessToken, err := ss.jwtService.CreateToken(authData)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.NewTokenResp(accessToken, refreshToken, rawFingerprint), nil
}

// RevokeSession deletes the session and all associated refresh tokens
func (ss *sessionService) RevokeSession(ctx context.Context, sessionID string) error {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.RevokeSession")
	defer span.End()

	return ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := ss.refreshTokens.DeleteBySession(ctx, sessionID); err != nil {
			return err
		}
		return ss.sessions.Delete(ctx, sessionID)
	})
}

// createSession creates a new session with initial refresh token
func (ss *sessionService) createSession(ctx context.Context, userID string, refreshTokenTTL time.Duration) (auth.Session, string, error) {
	var session auth.Session
	var refreshToken string

	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		var err error
		session, err = ss.sessions.Create(ctx, userID)
		if err != nil {
			return err
		}

		// Create initial refresh token
		expiresAt := time.Now().Add(refreshTokenTTL)
		refreshToken, err = ss.createRefreshToken(ctx, session.ID, expiresAt)
		if err != nil {
			return err
		}

		return nil
	})

	return session, refreshToken, err
}

// createRefreshToken issues a new refresh token for a session
func (ss *sessionService) createRefreshToken(ctx context.Context, sessionID string, expiresAt time.Time) (string, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.createRefreshToken")
	defer span.End()

	token, tokenHash, err := ss.generateRefreshToken()
	if err != nil {
		return "", err
	}

	if err = ss.refreshTokens.Create(ctx, sessionID, tokenHash, expiresAt); err != nil {
		return "", err
	}

	return token, nil
}

// generateRefreshToken creates a cryptographically secure random token
func (ss *sessionService) generateRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", ungerr.Wrap(err, "error generating random bytes")
	}

	token := hex.EncodeToString(bytes)

	return token, ss.hashToken(token), nil
}

func (ss *sessionService) GetByID(ctx context.Context, id string) (auth.Session, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.GetByID")
	defer span.End()

	session, err := ss.sessions.GetByID(ctx, id)
	if err != nil {
		return auth.Session{}, err
	}
	if session.IsZero() {
		return auth.Session{}, ungerr.UnauthorizedError("session not found")
	}
	return session, nil
}

// rotateRefreshToken safely rotates a refresh token with reuse detection
func (ss *sessionService) rotateRefreshToken(ctx context.Context, oldToken string) (string, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.rotateRefreshToken")
	defer span.End()

	var newToken string

	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		oldRefreshToken, err := ss.getRefreshToken(ctx, oldToken)
		if err != nil {
			return err
		}

		// Reject expired tokens
		if time.Now().After(oldRefreshToken.ExpiresAt) {
			return ungerr.UnauthorizedError("refresh token expired")
		}

		// Token is still valid — rotate it
		session, err := ss.GetByID(ctx, oldRefreshToken.SessionID)
		if err != nil {
			return err
		}

		// Delete the old refresh token
		if err = ss.refreshTokens.Delete(ctx, oldRefreshToken.SessionID, oldRefreshToken.TokenHash); err != nil {
			return err
		}

		// Session stays active; issue new refresh token
		expiresAt := time.Now().Add(ss.refreshTokenTTL)
		newToken, err = ss.createRefreshToken(ctx, session.ID, expiresAt)
		if err != nil {
			return err
		}

		// Touch session to update LastUsedAt
		if err = ss.sessions.Touch(ctx, session.ID); err != nil {
			return err
		}

		return nil
	})

	return newToken, err
}

// getRefreshToken looks up a refresh token by its raw value (hashed comparison)
func (ss *sessionService) getRefreshToken(ctx context.Context, rawToken string) (auth.RefreshToken, error) {
	ctx, span := otel.Tracer.Start(ctx, "SessionService.getRefreshToken")
	defer span.End()

	rt, err := ss.refreshTokens.FindByHash(ctx, ss.hashToken(rawToken))
	if err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return auth.RefreshToken{}, ungerr.UnauthorizedError("refresh token not found")
		}
		return auth.RefreshToken{}, err
	}
	return rt, nil
}

func (ss *sessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (sessionService) generateFingerprint() (raw string, hash string) {
	b := make([]byte, 32)
	rand.Read(b)
	raw = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(raw))
	hash = hex.EncodeToString(h[:])
	return
}
