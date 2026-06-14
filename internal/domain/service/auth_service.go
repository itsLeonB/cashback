package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service/auth"
	"github.com/itsLeonB/ungerr"
)

type authServiceImpl struct {
	hashService      auth.HashService
	jwtService       auth.JWTService
	transactor       auth.Transactor
	users            auth.UserStore
	resets           auth.ResetTokenStore
	mailSvc          auth.MailService
	verificationURL  string
	resetPasswordURL string
	sessionSvc       SessionService
	sessionCache     auth.SessionCache
	hooks            AuthHooks
}

func NewAuthService(
	jwtService auth.JWTService,
	transactor auth.Transactor,
	users auth.UserStore,
	resets auth.ResetTokenStore,
	mailSvc auth.MailService,
	verificationURL string,
	resetPasswordURL string,
	hashService auth.HashService,
	sessionSvc SessionService,
	sessionCache auth.SessionCache,
	hooks AuthHooks,
) AuthService {
	return &authServiceImpl{
		hashService:      hashService,
		jwtService:       jwtService,
		transactor:       transactor,
		users:            users,
		resets:           resets,
		mailSvc:          mailSvc,
		verificationURL:  verificationURL,
		resetPasswordURL: resetPasswordURL,
		sessionSvc:       sessionSvc,
		sessionCache:     sessionCache,
		hooks:            hooks,
	}
}

func (as *authServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (dto.RegisterResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.Register")
	defer span.End()

	isVerified, err := as.executeRegistration(ctx, req)
	if err != nil {
		return dto.RegisterResponse{}, err
	}

	msg := "check your email to confirm your registration"
	if isVerified {
		msg = "success registering, please login"
	}

	return dto.RegisterResponse{
		Message: msg,
	}, nil
}

func (as *authServiceImpl) executeRegistration(ctx context.Context, request dto.RegisterRequest) (bool, error) {
	isVerified := as.verificationURL == ""
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := as.users.FindByEmail(ctx, request.Email)
		if err != nil && !errors.Is(err, auth.ErrUserNotFound) {
			return err
		}
		if !user.IsZero() {
			return ungerr.ConflictError(fmt.Sprintf("email %s already exists", request.Email))
		}

		hash, err := as.hashService.Hash(request.Password)
		if err != nil {
			return err
		}

		name := util.GetNameFromEmail(request.Email)
		newUser, err := as.users.Create(ctx, request.Email, hash)
		if err != nil {
			return err
		}
		if isVerified {
			_, err = as.users.SetVerified(ctx, newUser.ID, name, "")
			return err
		}

		return as.sendVerificationMail(ctx, newUser, as.verificationURL, request.Slug)
	})
	return isVerified, err
}

func (as *authServiceImpl) sendVerificationMail(ctx context.Context, user auth.User, verificationURL string, slug string) error {
	claims := map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(30 * time.Minute).Unix(),
	}
	if slug != "" {
		claims["slug"] = slug
	}

	token, err := as.jwtService.CreateToken(claims)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?token=%s", verificationURL, token)

	name := util.GetNameFromEmail(user.Email)
	err = as.mailSvc.SendVerification(ctx, user.Email, name, url)
	return err
}

func (as *authServiceImpl) InternalLogin(ctx context.Context, req dto.InternalLoginRequest) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.InternalLogin")
	defer span.End()

	user, err := as.users.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
		}
		return dto.TokenResponse{}, err
	}
	if !user.Verified {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	ok, err := as.hashService.Verify(user.PasswordHash, req.Password)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if !ok {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	return as.sessionSvc.CreateTokenAndSession(ctx, user)
}

func (as *authServiceImpl) VerifyToken(ctx context.Context, token string, fingerprint string) (bool, map[string]any, error) {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.VerifyToken")
	defer span.End()

	claims, err := as.jwtService.VerifyToken(token)
	if err != nil {
		return false, nil, err
	}

	// Verify fingerprint
	expectedHash, fgpExists := claims.Data[appconstant.ContextFingerprint.String()]
	if !fgpExists {
		return false, nil, ungerr.UnauthorizedError("missing fingerprint claim")
	}
	expectedHashStr, ok := expectedHash.(string)
	if !ok {
		return false, nil, ungerr.UnauthorizedError("invalid fingerprint claim type")
	}
	hash := sha256.Sum256([]byte(fingerprint))
	if hex.EncodeToString(hash[:]) != expectedHashStr {
		return false, nil, ungerr.UnauthorizedError("invalid token fingerprint")
	}

	tokenUserId, exists := claims.Data[appconstant.ContextUserID.String()]
	if !exists {
		return false, nil, ungerr.Unknown("missing user ID from token")
	}
	stringUserID, ok := tokenUserId.(string)
	if !ok {
		return false, nil, ungerr.Unknown("error asserting userID, is not a string")
	}

	// Extract session_id from token
	sessionIDStr, exists := claims.Data[appconstant.ContextSessionID.String()]
	if !exists {
		return false, nil, ungerr.Unknown("missing session ID from token")
	}
	sessionIDString, ok := sessionIDStr.(string)
	if !ok {
		return false, nil, ungerr.Unknown("error asserting sessionID, is not a string")
	}

	var loadErr error
	cachedUserID, hit := as.sessionCache.Get(sessionIDString, func(_ string) (string, bool) {
		session, err := as.sessionSvc.GetByID(ctx, sessionIDString)
		if err != nil {
			loadErr = err
			return "", false
		}
		return session.UserID, true
	})
	if loadErr != nil {
		return false, nil, loadErr
	}
	if !hit {
		return false, nil, ungerr.UnauthorizedError("session is not found")
	}
	if cachedUserID != stringUserID {
		return false, nil, ungerr.UnauthorizedError("session does not belong to user")
	}

	// Parse UUID string fields back to uuid.UUID for handlers
	result := make(map[string]any, len(claims.Data))
	for k, v := range claims.Data {
		if s, ok := v.(string); ok {
			if uid, err := uuid.Parse(s); err == nil {
				result[k] = uid
				continue
			}
		}
		result[k] = v
	}

	return true, result, nil
}

func (as *authServiceImpl) VerifyRegistration(ctx context.Context, token string) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.VerifyRegistration")
	defer span.End()

	var response dto.TokenResponse
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		claims, err := as.jwtService.VerifyToken(token)
		if err != nil {
			return err
		}
		id, ok := claims.Data["id"].(string)
		if !ok {
			return ungerr.Unknown("error asserting id, is not a string")
		}
		email, ok := claims.Data["email"].(string)
		if !ok {
			return ungerr.Unknown("error asserting email, is not a string")
		}
		exp, ok := claims.Data["exp"].(float64)
		if !ok {
			return ungerr.Unknown("error asserting exp, is not an float64")
		}
		unixTime := int64(exp)
		if time.Now().Unix() > unixTime {
			return ungerr.UnauthorizedError("token has expired")
		}

		user, err := as.users.SetVerified(ctx, id, util.GetNameFromEmail(email), "")
		if err != nil {
			return err
		}

		if err := as.hooks.CallAfterEmailVerified(ctx, user.ID, user.ProfileID, claims.Data); err != nil {
			return err
		}

		response, err = as.sessionSvc.CreateTokenAndSession(ctx, user)
		return err
	})
	return response, err
}

func (as *authServiceImpl) SendPasswordReset(ctx context.Context, email string) error {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.SendPasswordReset")
	defer span.End()

	return as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := as.users.FindByEmail(ctx, email)
		if err != nil {
			if errors.Is(err, auth.ErrUserNotFound) {
				return nil
			}
			return err
		}
		if !user.Verified {
			return nil
		}

		selector, err := generateToken()
		if err != nil {
			return err
		}
		verifier, err := generateToken()
		if err != nil {
			return err
		}
		verifierHash := hashVerifier(verifier)
		expiresAt := time.Now().Add(1 * time.Hour)

		err = as.resets.Create(ctx, user.ID, selector, verifierHash, expiresAt)
		if err != nil {
			return err
		}

		return as.sendResetPasswordMail(ctx, user, as.resetPasswordURL, selector, verifier)
	})
}

func (as *authServiceImpl) sendResetPasswordMail(ctx context.Context, user auth.User, resetURL, selector, verifier string) error {
	claims := map[string]any{
		"id":       user.ID,
		"email":    user.Email,
		"selector": selector,
		"verifier": verifier,
	}

	token, err := as.jwtService.CreateToken(claims)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?token=%s", resetURL, token)

	name := util.GetNameFromEmail(user.Email)
	return as.mailSvc.SendPasswordReset(ctx, user.Email, name, url)
}

func (as *authServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) (dto.TokenResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.ResetPassword")
	defer span.End()

	var response dto.TokenResponse
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		claims, err := as.jwtService.VerifyToken(token)
		if err != nil {
			return err
		}
		id, ok := claims.Data["id"].(string)
		if !ok {
			return ungerr.Unknown("error asserting id, is not a string")
		}
		email, ok := claims.Data["email"].(string)
		if !ok {
			return ungerr.Unknown("error asserting email, is not a string")
		}
		selector, ok := claims.Data["selector"].(string)
		if !ok {
			return ungerr.Unknown("error asserting selector, is not a string")
		}
		verifier, ok := claims.Data["verifier"].(string)
		if !ok {
			return ungerr.Unknown("error asserting verifier, is not a string")
		}

		// Validate selector/verifier
		resetToken, err := as.resets.FindBySelector(ctx, selector)
		if err != nil {
			if errors.Is(err, auth.ErrTokenNotFound) {
				return ungerr.UnauthorizedError("token is invalid")
			}
			return err
		}
		if resetToken.ExpiresAt.Before(time.Now()) {
			return ungerr.UnauthorizedError("token has expired")
		}

		verifierHash := hashVerifier(verifier)
		if subtle.ConstantTimeCompare([]byte(resetToken.VerifierHash), []byte(verifierHash)) != 1 {
			return ungerr.UnauthorizedError("token is invalid")
		}

		hashedPassword, err := as.hashService.Hash(newPassword)
		if err != nil {
			return err
		}

		err = as.users.UpdatePassword(ctx, id, hashedPassword)
		if err != nil {
			return err
		}

		err = as.resets.DeleteByUser(ctx, id)
		if err != nil {
			return err
		}

		user := auth.User{
			ID:    id,
			Email: email,
		}
		response, err = as.sessionSvc.CreateTokenAndSession(ctx, user)
		return err
	})
	return response, err
}

// Logout revokes the current session and all its refresh tokens
func (as *authServiceImpl) Logout(ctx context.Context, sessionID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "AuthService.Logout")
	defer span.End()

	sid := sessionID.String()

	if err := as.hooks.CallBeforeLogout(ctx, sid); err != nil {
		logger.Error(err)
	}

	as.sessionCache.Delete(sid)

	// Revoke session and refresh tokens
	return as.sessionSvc.RevokeSession(ctx, sid)
}

func (as *authServiceImpl) Shutdown() error {
	return as.sessionCache.Shutdown()
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", ungerr.Wrap(err, "error generating token")
	}
	return hex.EncodeToString(b), nil
}

func hashVerifier(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return hex.EncodeToString(h[:])
}
