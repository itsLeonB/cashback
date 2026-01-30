package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/mail"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/sekure"
	"github.com/itsLeonB/ungerr"
)

type authServiceImpl struct {
	hashService      sekure.HashService
	jwtService       sekure.JWTService
	transactor       crud.Transactor
	userSvc          UserService
	mailSvc          mail.MailService
	verificationURL  string
	resetPasswordURL string
	oAuthSvc         OAuthService
	sessionRepo      crud.Repository[users.Session]
	refreshTokenRepo crud.Repository[users.RefreshToken]
	pushSvc          PushNotificationService
}

func NewAuthService(
	jwtService sekure.JWTService,
	transactor crud.Transactor,
	userSvc UserService,
	mailSvc mail.MailService,
	verificationURL string,
	resetPasswordURL string,
	oAuthSvc OAuthService,
	hashCost int,
	sessionRepo crud.Repository[users.Session],
	refreshTokenRepo crud.Repository[users.RefreshToken],
	pushSvc PushNotificationService,
) AuthService {
	return &authServiceImpl{
		sekure.NewHashService(hashCost),
		jwtService,
		transactor,
		userSvc,
		mailSvc,
		verificationURL,
		resetPasswordURL,
		oAuthSvc,
		sessionRepo,
		refreshTokenRepo,
		pushSvc,
	}
}

func (as *authServiceImpl) Register(ctx context.Context, req dto.RegisterRequest) (dto.RegisterResponse, error) {
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
		existingUser, err := as.userSvc.FindByEmail(ctx, request.Email)
		if err != nil {
			return err
		}
		if !existingUser.IsZero() {
			return ungerr.ConflictError(fmt.Sprintf("email %s already exists", request.Email))
		}

		hash, err := as.hashService.Hash(request.Password)
		if err != nil {
			return err
		}

		newUserReq := dto.NewUserRequest{
			Email:     request.Email,
			Password:  hash,
			Name:      getNameFromEmail(request.Email),
			VerifyNow: isVerified,
		}

		user, err := as.userSvc.CreateNew(ctx, newUserReq)
		if err != nil {
			return err
		}
		if isVerified {
			return nil
		}

		return as.sendVerificationMail(ctx, user, as.verificationURL)
	})
	return isVerified, err
}

func (as *authServiceImpl) sendVerificationMail(ctx context.Context, user users.User, verificationURL string) error {
	claims := map[string]any{
		"id":    user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(30 * time.Minute).Unix(),
	}

	token, err := as.jwtService.CreateToken(claims)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?token=%s", verificationURL, token)

	mailMsg := mail.MailMessage{
		RecipientMail: user.Email,
		RecipientName: getNameFromEmail(user.Email),
		Subject:       "Verify your email",
		TextContent:   "Please verify your email by clicking the following link:\n\n" + url,
	}

	return as.mailSvc.Send(ctx, mailMsg)
}

func getNameFromEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) < 2 || parts[0] == "" {
		return ""
	}
	localPart := parts[0]

	re := regexp.MustCompile(`[a-zA-Z]+`)
	matches := re.FindAllString(localPart, -1)
	if len(matches) > 0 {
		name := matches[0]
		return ezutil.Capitalize(name)
	}

	return ""
}

func (as *authServiceImpl) InternalLogin(ctx context.Context, req dto.InternalLoginRequest) (dto.LoginResponse, error) {
	user, err := as.userSvc.FindByEmail(ctx, req.Email)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if user.IsZero() {
		return dto.LoginResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}
	if !user.IsVerified() {
		return dto.LoginResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	ok, err := as.hashService.CheckHash(user.Password, req.Password)
	if err != nil {
		return dto.LoginResponse{}, err
	}
	if !ok {
		return dto.LoginResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	// Create session with refresh token
	session, refreshToken, err := as.createSession(ctx, user.ID, "", 30*24*time.Hour) // 30 day refresh token
	if err != nil {
		return dto.LoginResponse{}, err
	}

	// Create access token
	authData := mapper.UserToAuthData(user, session)
	accessToken, err := as.jwtService.CreateToken(authData)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.NewBearerTokenWithRefreshResp(accessToken, refreshToken), nil
}

func (as *authServiceImpl) VerifyToken(ctx context.Context, token string) (bool, map[string]any, error) {
	claims, err := as.jwtService.VerifyToken(token)
	if err != nil {
		return false, nil, err
	}

	tokenUserId, exists := claims.Data[appconstant.ContextUserID.String()]
	if !exists {
		return false, nil, ungerr.Unknown("missing user ID from token")
	}
	stringUserID, ok := tokenUserId.(string)
	if !ok {
		return false, nil, ungerr.Unknown("error asserting userID, is not a string")
	}
	userID, err := ezutil.Parse[uuid.UUID](stringUserID)
	if err != nil {
		return false, nil, err
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
	sessionID, err := ezutil.Parse[uuid.UUID](sessionIDString)
	if err != nil {
		return false, nil, err
	}

	user, err := as.userSvc.GetByID(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	return true, map[string]any{
		appconstant.ContextProfileID.String(): user.Profile.ID,
		appconstant.ContextSessionID.String(): sessionID,
	}, nil
}

func (as *authServiceImpl) GetOAuth2URL(ctx context.Context, provider string) (string, error) {
	return as.oAuthSvc.GetOAuthURL(ctx, provider)
}

func (as *authServiceImpl) OAuth2Login(ctx context.Context, provider, code, state string) (dto.LoginResponse, error) {
	return as.oAuthSvc.HandleOAuthCallback(ctx, dto.OAuthCallbackData{
		Provider: provider,
		Code:     code,
		State:    state,
	})
}

func (as *authServiceImpl) VerifyRegistration(ctx context.Context, token string) (dto.LoginResponse, error) {
	var response dto.LoginResponse
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		claims, err := as.jwtService.VerifyToken(token)
		if err != nil {
			return err
		}
		id, ok := claims.Data["id"].(string)
		if !ok {
			return ungerr.Unknown("error asserting id, is not a string")
		}
		userID, err := ezutil.Parse[uuid.UUID](id)
		if err != nil {
			return err
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

		user, err := as.userSvc.Verify(ctx, userID, email, getNameFromEmail(email), "")
		if err != nil {
			return err
		}

		response, err = as.oAuthSvc.CreateLoginResponse(user, users.Session{})
		return err
	})
	return response, err
}

func (as *authServiceImpl) SendPasswordReset(ctx context.Context, email string) error {
	return as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		user, err := as.userSvc.FindByEmail(ctx, email)
		if err != nil {
			return err
		}
		if user.IsZero() || !user.IsVerified() {
			return nil
		}

		resetToken, err := as.userSvc.GeneratePasswordResetToken(ctx, user.ID)
		if err != nil {
			return err
		}

		return as.sendResetPasswordMail(ctx, user, as.resetPasswordURL, resetToken)
	})
}

func (as *authServiceImpl) sendResetPasswordMail(ctx context.Context, user users.User, resetURL, resetToken string) error {
	claims := map[string]any{
		"id":          user.ID,
		"email":       user.Email,
		"reset_token": resetToken,
	}

	token, err := as.jwtService.CreateToken(claims)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s?token=%s", resetURL, token)

	mailMsg := mail.MailMessage{
		RecipientMail: user.Email,
		RecipientName: user.Profile.Name,
		Subject:       "Reset your password",
		TextContent:   "You have requested to reset your password.\nIf this is not you, ignore this mail.\nPlease reset your password by clicking the following link:\n\n" + url,
	}

	return as.mailSvc.Send(ctx, mailMsg)
}

func (as *authServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) (dto.LoginResponse, error) {
	var response dto.LoginResponse
	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		claims, err := as.jwtService.VerifyToken(token)
		if err != nil {
			return err
		}
		id, ok := claims.Data["id"].(string)
		if !ok {
			return ungerr.Unknown("error asserting id, is not a string")
		}
		userID, err := ezutil.Parse[uuid.UUID](id)
		if err != nil {
			return err
		}
		email, ok := claims.Data["email"].(string)
		if !ok {
			return ungerr.Unknown("error asserting email, is not a string")
		}
		resetToken, ok := claims.Data["reset_token"].(string)
		if !ok {
			return ungerr.Unknown("error asserting reset_token, is not a string")
		}

		hashedPassword, err := as.hashService.Hash(newPassword)
		if err != nil {
			return err
		}

		user, err := as.userSvc.ResetPassword(ctx, userID, email, resetToken, hashedPassword)
		if err != nil {
			return err
		}

		response, err = as.oAuthSvc.CreateLoginResponse(user, users.Session{})
		return err
	})
	return response, err
}

// generateRefreshToken creates a cryptographically secure random token
func (as *authServiceImpl) generateRefreshToken() (string, string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", ungerr.Wrap(err, "error generating random bytes")
	}

	token := hex.EncodeToString(bytes)

	return token, hashToken(token), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// createRefreshToken issues a new refresh token for a session
func (as *authServiceImpl) createRefreshToken(ctx context.Context, sessionID uuid.UUID, expiresAt time.Time) (string, error) {
	token, tokenHash, err := as.generateRefreshToken()
	if err != nil {
		return "", err
	}

	refreshToken := users.RefreshToken{
		SessionID: sessionID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	}

	_, err = as.refreshTokenRepo.Insert(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return token, nil
}

// rotateRefreshToken safely rotates a refresh token with reuse detection
func (as *authServiceImpl) rotateRefreshToken(ctx context.Context, oldToken string) (string, error) {
	var newToken string

	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		oldRefreshToken, err := as.getRefreshToken(ctx, oldToken)
		if err != nil {
			return err
		}

		// Check if token is expired
		if time.Now().After(oldRefreshToken.ExpiresAt) {
			return ungerr.UnauthorizedError("refresh token expired")
		}

		session, err := as.getSessionByID(ctx, oldRefreshToken.SessionID)
		if err != nil {
			return err
		}

		// Delete the old refresh token (hard delete for rotation)
		if err = as.refreshTokenRepo.Delete(ctx, oldRefreshToken); err != nil {
			return err
		}

		// Create new refresh token with same expiry duration
		duration := oldRefreshToken.ExpiresAt.Sub(oldRefreshToken.CreatedAt)
		newExpiresAt := time.Now().Add(duration)

		newToken, err = as.createRefreshToken(ctx, session.ID, newExpiresAt)
		if err != nil {
			return err
		}

		// Update session last used time
		session.LastUsedAt = time.Now()
		session.UpdatedAt = time.Now()
		_, err = as.sessionRepo.Update(ctx, session)
		return err
	})

	return newToken, err
}

func (as *authServiceImpl) getRefreshToken(ctx context.Context, token string) (users.RefreshToken, error) {
	spec := crud.Specification[users.RefreshToken]{}
	spec.Model.TokenHash = hashToken(token)
	refreshToken, err := as.refreshTokenRepo.FindFirst(ctx, spec)
	if err != nil {
		return users.RefreshToken{}, err
	}
	if refreshToken.IsZero() {
		return users.RefreshToken{}, ungerr.UnauthorizedError("invalid refresh token")
	}
	return refreshToken, nil
}

// revokeSession deletes the session and all associated refresh tokens
func (as *authServiceImpl) revokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Delete all refresh tokens for this session
		spec := crud.Specification[users.RefreshToken]{}
		spec.Model.SessionID = sessionID
		refreshTokens, err := as.refreshTokenRepo.FindAll(ctx, spec)
		if err != nil {
			return err
		}

		if err = as.refreshTokenRepo.DeleteMany(ctx, refreshTokens); err != nil {
			return err
		}

		// Delete the session
		session, err := as.findSessionByID(ctx, sessionID)
		if err != nil {
			return err
		}
		if session.IsZero() {
			return nil
		}
		return as.sessionRepo.Delete(ctx, session)
	})
}

func (as *authServiceImpl) getSessionByID(ctx context.Context, id uuid.UUID) (users.Session, error) {
	session, err := as.findSessionByID(ctx, id)
	if err != nil {
		return users.Session{}, err
	}
	if session.IsZero() {
		return users.Session{}, ungerr.UnauthorizedError("session not found")
	}
	return session, nil
}

// createSession creates a new session with initial refresh token
func (as *authServiceImpl) createSession(ctx context.Context, userID uuid.UUID, deviceID string, refreshTokenTTL time.Duration) (users.Session, string, error) {
	var session users.Session
	var refreshToken string

	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		// Create session
		session := users.Session{
			UserID:     userID,
			LastUsedAt: time.Now(),
		}

		if deviceID != "" {
			session.DeviceID = sql.NullString{
				String: deviceID,
				Valid:  true,
			}
		}

		insertedSession, err := as.sessionRepo.Insert(ctx, session)
		if err != nil {
			return err
		}

		// Create initial refresh token
		expiresAt := time.Now().Add(refreshTokenTTL)
		refreshToken, err = as.createRefreshToken(ctx, insertedSession.ID, expiresAt)
		if err != nil {
			return err
		}

		session = insertedSession
		return nil
	})

	return session, refreshToken, err
}

// RefreshToken validates and rotates a refresh token, issuing new access and refresh tokens
func (as *authServiceImpl) RefreshToken(ctx context.Context, request dto.RefreshTokenRequest) (dto.RefreshTokenResponse, error) {
	var response dto.RefreshTokenResponse

	err := as.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		refreshToken, err := as.getRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			return err
		}

		session, err := as.getSessionByID(ctx, refreshToken.SessionID)
		if err != nil {
			return err
		}

		// Get user for JWT claims
		user, err := as.userSvc.GetByID(ctx, session.UserID)
		if err != nil {
			return err
		}

		// Rotate the refresh token (this validates expiry and deletes old token)
		newRefreshToken, err := as.rotateRefreshToken(ctx, request.RefreshToken)
		if err != nil {
			return err
		}

		claims := mapper.UserToAuthData(user, session)

		accessToken, err := as.jwtService.CreateToken(claims)
		if err != nil {
			return err
		}

		response = dto.NewRefreshTokenResp(accessToken, newRefreshToken)
		return nil
	})

	return response, err
}

func (as *authServiceImpl) findSessionByID(ctx context.Context, id uuid.UUID) (users.Session, error) {
	spec := crud.Specification[users.Session]{}
	spec.Model.ID = id
	return as.sessionRepo.FindFirst(ctx, spec)
}

// Logout revokes the current session and all its refresh tokens
func (as *authServiceImpl) Logout(ctx context.Context, sessionID uuid.UUID) error {
	// Clean up push subscriptions for this session (failure must not block logout)
	if err := as.pushSvc.UnsubscribeBySession(ctx, sessionID); err != nil {
		logger.Error(err)
	}

	// Revoke session and refresh tokens
	return as.revokeSession(ctx, sessionID)
}
