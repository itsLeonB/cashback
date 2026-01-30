package dto

type RegisterRequest struct {
	Email                string `json:"email" binding:"required,email,min=3"`
	Password             string `json:"password" binding:"required,eqfield=PasswordConfirmation"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"required"`
}

type InternalLoginRequest struct {
	Email    string `json:"email" binding:"required,email,min=3"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type LoginResponse struct {
	Type         string `json:"type"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type RefreshTokenResponse struct {
	Type         string `json:"type"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type SendPasswordResetRequest struct {
	Email string `json:"email" binding:"required,email,min=3"`
}

type ResetPasswordRequest struct {
	Token                string `json:"token" binding:"required,min=3"`
	Password             string `json:"password" binding:"required,eqfield=PasswordConfirmation"`
	PasswordConfirmation string `json:"passwordConfirmation" binding:"required"`
}

type OAuthCallbackData struct {
	Provider string `validate:"required,min=1"`
	Code     string `validate:"required,min=1"`
	State    string `validate:"required,min=1"`
}

func NewBearerTokenResp(token string) LoginResponse {
	return LoginResponse{
		Type:  "Bearer",
		Token: token,
	}
}

func NewBearerTokenWithRefreshResp(token, refreshToken string) LoginResponse {
	return LoginResponse{
		Type:         "Bearer",
		Token:        token,
		RefreshToken: refreshToken,
	}
}

func NewRefreshTokenResp(token, refreshToken string) RefreshTokenResponse {
	return RefreshTokenResponse{
		Type:         "Bearer",
		Token:        token,
		RefreshToken: refreshToken,
	}
}
