package auth

import "errors"

var (
	ErrUserNotFound       = errors.New("auth: user not found")
	ErrUserExists         = errors.New("auth: user already exists")
	ErrSessionNotFound    = errors.New("auth: session not found")
	ErrTokenNotFound      = errors.New("auth: token not found")
	ErrTokenExpired       = errors.New("auth: token expired")
	ErrTokenInvalid       = errors.New("auth: token invalid")
	ErrInvalidCredentials = errors.New("auth: invalid credentials")
	ErrNotVerified        = errors.New("auth: user not verified")
	ErrProviderDisabled   = errors.New("auth: provider disabled")
)
