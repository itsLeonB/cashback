package service

import (
	"context"
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
	pushSvc          PushNotificationService
	sessionSvc       SessionService
}

func NewAuthService(
	jwtService sekure.JWTService,
	transactor crud.Transactor,
	userSvc UserService,
	mailSvc mail.MailService,
	verificationURL string,
	resetPasswordURL string,
	hashCost int,
	pushSvc PushNotificationService,
	sessionSvc SessionService,
) AuthService {
	return &authServiceImpl{
		sekure.NewHashService(hashCost),
		jwtService,
		transactor,
		userSvc,
		mailSvc,
		verificationURL,
		resetPasswordURL,
		pushSvc,
		sessionSvc,
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

func (as *authServiceImpl) InternalLogin(ctx context.Context, req dto.InternalLoginRequest) (dto.TokenResponse, error) {
	user, err := as.userSvc.FindByEmail(ctx, req.Email)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if user.IsZero() {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}
	if !user.IsVerified() {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	ok, err := as.hashService.CheckHash(user.Password, req.Password)
	if err != nil {
		return dto.TokenResponse{}, err
	}
	if !ok {
		return dto.TokenResponse{}, ungerr.NotFoundError(appconstant.ErrAuthUnknownCredentials)
	}

	return as.sessionSvc.CreateTokenAndSession(ctx, user)
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

	if _, err = as.sessionSvc.GetByID(ctx, sessionID); err != nil {
		return false, nil, ungerr.UnauthorizedError("session is not found")
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

func (as *authServiceImpl) VerifyRegistration(ctx context.Context, token string) (dto.TokenResponse, error) {
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

		response, err = as.sessionSvc.CreateTokenAndSession(ctx, user)
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

func (as *authServiceImpl) ResetPassword(ctx context.Context, token, newPassword string) (dto.TokenResponse, error) {
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

		response, err = as.sessionSvc.CreateTokenAndSession(ctx, user)
		return err
	})
	return response, err
}

// Logout revokes the current session and all its refresh tokens
func (as *authServiceImpl) Logout(ctx context.Context, sessionID uuid.UUID) error {
	// Clean up push subscriptions for this session (failure must not block logout)
	if err := as.pushSvc.UnsubscribeBySession(ctx, sessionID); err != nil {
		logger.Error(err)
	}

	// Revoke session and refresh tokens
	return as.sessionSvc.RevokeSession(ctx, sessionID)
}
