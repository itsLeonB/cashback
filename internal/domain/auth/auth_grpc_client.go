package auth

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/cocoon-protos/gen/go/auth/v1"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/ungerr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthClient interface {
	Register(ctx context.Context, req RegisterRequest) (bool, error)
	InternalLogin(ctx context.Context, req InternalLoginRequest) (LoginResponse, error)
	OAuth2Login(ctx context.Context, req OAuthLoginRequest) (LoginResponse, error)
	VerifyToken(ctx context.Context, token string) (bool, map[string]any, error)
	GetOAuth2URL(ctx context.Context, provider string) (string, error)
	VerifyRegistration(ctx context.Context, token string) (LoginResponse, error)
	SendPasswordReset(ctx context.Context, resetURL, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) (LoginResponse, error)
}

type authClient struct {
	validate *validator.Validate
	client   auth.AuthServiceClient
}

func NewAuthClient(validate *validator.Validate, conn *grpc.ClientConn) AuthClient {
	if validate == nil {
		panic("validator is nil")
	}

	return &authClient{
		validate: validate,
		client:   auth.NewAuthServiceClient(conn),
	}
}

func (ac *authClient) Register(ctx context.Context, req RegisterRequest) (bool, error) {
	if err := ac.validate.Struct(req); err != nil {
		return false, err
	}

	request := &auth.RegisterRequest{
		Email:                req.Email,
		Password:             req.Password,
		PasswordConfirmation: req.PasswordConfirmation,
		VerificationUrl:      req.VerificationURL,
	}

	response, err := ac.client.Register(ctx, request)
	if err != nil {
		return false, err
	}

	return response.GetIsVerified(), nil
}

func (ac *authClient) InternalLogin(ctx context.Context, req InternalLoginRequest) (LoginResponse, error) {
	if err := ac.validate.Struct(req); err != nil {
		return LoginResponse{}, err
	}

	request := auth.LoginRequest{
		LoginMethod: &auth.LoginRequest_InternalRequest{
			InternalRequest: &auth.InternalLoginRequest{
				Email:    req.Email,
				Password: req.Password,
			},
		},
	}

	response, err := ac.client.Login(ctx, &request)
	if err != nil {
		return LoginResponse{}, err
	}

	return LoginResponse{
		Type:  response.GetType(),
		Token: response.GetToken(),
	}, nil
}

func (ac *authClient) OAuth2Login(ctx context.Context, req OAuthLoginRequest) (LoginResponse, error) {
	if err := ac.validate.Struct(req); err != nil {
		return LoginResponse{}, err
	}

	request := auth.LoginRequest{
		LoginMethod: &auth.LoginRequest_Oauth2Request{
			Oauth2Request: &auth.OAuth2LoginRequest{
				Provider: req.Provider,
				Code:     req.Code,
				State:    req.State,
			},
		},
	}

	response, err := ac.client.Login(ctx, &request)
	if err != nil {
		return LoginResponse{}, err
	}

	return fromLoginResponseProto(response), nil
}

func (ac *authClient) VerifyToken(ctx context.Context, token string) (bool, map[string]any, error) {
	request := &auth.VerifyTokenRequest{Token: token}

	data, err := ac.client.VerifyToken(ctx, request)
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			return false, nil, ungerr.UnauthorizedError("unauthorized")
		}
		return false, nil, err
	}

	return true, map[string]any{
		appconstant.ContextProfileID.String(): data.GetProfileId(),
	}, nil
}

func (ac *authClient) GetOAuth2URL(ctx context.Context, provider string) (string, error) {
	request := auth.GetOAuth2UrlRequest{Provider: provider}
	response, err := ac.client.GetOAuth2Url(ctx, &request)
	if err != nil {
		return "", err
	}

	return response.GetUrl(), nil
}

func (ac *authClient) VerifyRegistration(ctx context.Context, token string) (LoginResponse, error) {
	request := auth.VerifyRegistrationRequest{Token: token}

	response, err := ac.client.VerifyRegistration(ctx, &request)
	if err != nil {
		return LoginResponse{}, err
	}

	return fromLoginResponseProto(response), nil
}

func (ac *authClient) SendPasswordReset(ctx context.Context, resetURL, email string) error {
	request := auth.SendResetPasswordRequest{
		ResetUrl: resetURL,
		Email:    email,
	}

	_, err := ac.client.SendResetPassword(ctx, &request)
	return err
}

func (ac *authClient) ResetPassword(ctx context.Context, token, newPassword string) (LoginResponse, error) {
	request := auth.ResetPasswordRequest{
		Token:       token,
		NewPassword: newPassword,
	}

	response, err := ac.client.ResetPassword(ctx, &request)
	if err != nil {
		return LoginResponse{}, err
	}

	return fromLoginResponseProto(response), nil
}
