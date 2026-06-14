package auth

import (
	"context"
	"time"
)

// UserStore handles user persistence for auth operations.
// Implementations should return ErrUserNotFound when no user matches.
type UserStore interface {
	FindByID(ctx context.Context, userID string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Create(ctx context.Context, email, passwordHash string) (User, error)
	CreateOAuth(ctx context.Context, email, name, avatar string) (User, error)
	SetVerified(ctx context.Context, userID string, name, avatar string) (User, error)
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	Exists(ctx context.Context, userID string) error
}

// SessionStore handles session persistence.
type SessionStore interface {
	Create(ctx context.Context, userID string) (Session, error)
	GetByID(ctx context.Context, id string) (Session, error)
	Delete(ctx context.Context, id string) error
	Touch(ctx context.Context, id string) error
}

// RefreshTokenStore handles refresh token persistence.
type RefreshTokenStore interface {
	Create(ctx context.Context, sessionID, tokenHash string, expiresAt time.Time) error
	FindByHash(ctx context.Context, hash string) (RefreshToken, error)
	Delete(ctx context.Context, sessionID string, tokenHash string) error
	DeleteBySession(ctx context.Context, sessionID string) error
}

// OAuthAccountStore handles OAuth account linking.
type OAuthAccountStore interface {
	FindByProvider(ctx context.Context, provider, providerID string) (OAuthAccount, error)
	Link(ctx context.Context, userID, provider, providerID, email string) error
}

// ResetTokenStore handles password reset tokens using selector/verifier pattern.
// The store only persists the selector + hashed verifier; the raw verifier never
// touches the database.
type ResetTokenStore interface {
	Create(ctx context.Context, userID, selector, verifierHash string, expiresAt time.Time) error
	FindBySelector(ctx context.Context, selector string) (ResetToken, error)
	DeleteByUser(ctx context.Context, userID string) error
}
