package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/sekure"
	"github.com/itsLeonB/ungerr"
)

type sessionService struct {
	jwtService       sekure.JWTService
	userSvc          UserService
	transactor       crud.Transactor
	sessionRepo      crud.Repository[users.Session]
	refreshTokenRepo crud.Repository[users.RefreshToken]
}

func NewSessionService(
	jwtService sekure.JWTService,
	userSvc UserService,
	transactor crud.Transactor,
	sessionRepo crud.Repository[users.Session],
	refreshTokenRepo crud.Repository[users.RefreshToken],
) *sessionService {
	return &sessionService{
		jwtService,
		userSvc,
		transactor,
		sessionRepo,
		refreshTokenRepo,
	}
}

// RefreshToken validates and rotates a refresh token, issuing new access and refresh tokens
func (ss *sessionService) RefreshToken(ctx context.Context, request dto.RefreshTokenRequest) (dto.TokenResponse, error) {
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
		if _, err := ss.userSvc.GetByID(ctx, session.UserID); err != nil {
			return err
		}

		// Rotate the refresh token (this validates expiry and deletes old token)
		newRefreshToken, err := ss.rotateRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			return err
		}

		claims := mapper.SessionToAuthData(session)

		accessToken, err := ss.jwtService.CreateToken(claims)
		if err != nil {
			return err
		}

		response = dto.NewTokenResp(accessToken, newRefreshToken)
		return nil
	})

	return response, err
}

func (ss *sessionService) CreateTokenAndSession(ctx context.Context, user users.User) (dto.TokenResponse, error) {
	// Create session with refresh token
	session, refreshToken, err := ss.createSession(ctx, user.ID, "", 30*24*time.Hour) // 30 day refresh token
	if err != nil {
		return dto.TokenResponse{}, err
	}

	// Create access token
	authData := mapper.SessionToAuthData(session)
	accessToken, err := ss.jwtService.CreateToken(authData)
	if err != nil {
		return dto.TokenResponse{}, err
	}

	return dto.NewTokenResp(accessToken, refreshToken), nil
}

// revokeSession deletes the session and all associated refresh tokens
func (ss *sessionService) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Delete all refresh tokens for this session
		spec := crud.Specification[users.RefreshToken]{}
		spec.Model.SessionID = sessionID
		refreshTokens, err := ss.refreshTokenRepo.FindAll(ctx, spec)
		if err != nil {
			return err
		}

		if err = ss.refreshTokenRepo.DeleteMany(ctx, refreshTokens); err != nil {
			return err
		}

		// Delete the session
		session, err := ss.findSessionByID(ctx, sessionID)
		if err != nil {
			return err
		}
		if session.IsZero() {
			return nil
		}
		return ss.sessionRepo.Delete(ctx, session)
	})
}

// createSession creates a new session with initial refresh token
func (ss *sessionService) createSession(ctx context.Context, userID uuid.UUID, deviceID string, refreshTokenTTL time.Duration) (users.Session, string, error) {
	var session users.Session
	var refreshToken string

	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create session
		session = users.Session{
			UserID:     userID,
			LastUsedAt: time.Now(),
		}

		if deviceID != "" {
			session.DeviceID = sql.NullString{
				String: deviceID,
				Valid:  true,
			}
		}

		insertedSession, err := ss.sessionRepo.Insert(ctx, session)
		if err != nil {
			return err
		}

		// Create initial refresh token
		expiresAt := time.Now().Add(refreshTokenTTL)
		refreshToken, err = ss.createRefreshToken(ctx, insertedSession.ID, expiresAt)
		if err != nil {
			return err
		}

		session = insertedSession
		return nil
	})

	return session, refreshToken, err
}

// createRefreshToken issues a new refresh token for a session
func (ss *sessionService) createRefreshToken(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) (string, error) {
	token, tokenHash, err := ss.generateRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := users.RefreshToken{
		SessionID: sessionID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	_, err = ss.refreshTokenRepo.Insert(ctx, refreshToken)
	if err != nil {
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

func (ss *sessionService) GetByID(ctx context.Context, id uuid.UUID) (users.Session, error) {
	session, err := ss.findSessionByID(ctx, id)
	if err != nil {
		return users.Session{}, err
	}
	if session.IsZero() {
		return users.Session{}, ungerr.UnauthorizedError("session not found")
	}
	return session, nil
}

func (ss *sessionService) findSessionByID(ctx context.Context, id uuid.UUID) (users.Session, error) {
	spec := crud.Specification[users.Session]{}
	spec.Model.ID = id
	return ss.sessionRepo.FindFirst(ctx, spec)
}

// rotateRefreshToken safely rotates a refresh token with reuse detection
func (ss *sessionService) rotateRefreshToken(ctx context.Context, oldToken string) (string, error) {
	var newToken string

	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		oldRefreshToken, err := ss.getRefreshToken(ctx, oldToken)
		if err != nil {
			return err
		}

		// Check if token is expired
		if time.Now().After(oldRefreshToken.ExpiresAt) {
			return ungerr.UnauthorizedError("refresh token expired")
		}

		session, err := ss.GetByID(ctx, oldRefreshToken.SessionID)
		if err != nil {
			return err
		}

		// Delete the old refresh token (hard delete for rotation)
		if err = ss.refreshTokenRepo.Delete(ctx, oldRefreshToken); err != nil {
			return err
		}

		// Create new refresh token with same expiry duration
		duration := oldRefreshToken.ExpiresAt.Sub(oldRefreshToken.CreatedAt)
		newExpiresAt := time.Now().Add(duration)

		newToken, err = ss.createRefreshToken(ctx, session.ID, newExpiresAt)
		if err != nil {
			return err
		}

		// Update session last used time
		session.LastUsedAt = time.Now()
		session.UpdatedAt = time.Now()
		_, err = ss.sessionRepo.Update(ctx, session)
		return err
	})

	return newToken, err
}

func (ss *sessionService) getRefreshToken(ctx context.Context, token string) (users.RefreshToken, error) {
	spec := crud.Specification[users.RefreshToken]{}
	spec.Model.TokenHash = ss.hashToken(token)
	refreshToken, err := ss.refreshTokenRepo.FindFirst(ctx, spec)
	if err != nil {
		return users.RefreshToken{}, err
	}
	if refreshToken.IsZero() {
		return users.RefreshToken{}, ungerr.UnauthorizedError("invalid refresh token")
	}
	return refreshToken, nil
}

func (sessionService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
