package service

import (
	"context"

	"github.com/google/uuid"
)

// AuthHooks holds optional callbacks that allow application-specific business
// logic to be injected into the authentication flow without coupling the auth
// service layer to concrete implementations. All fields are optional; nil
// fields are treated as no-ops.
type AuthHooks struct {
	// BeforeLogout runs before the session is revoked during logout. It is
	// called after the request is validated but before session deletion. Errors
	// returned by this hook are non-blocking and do not abort the logout flow.
	//
	// Use cases: unsubscribe push notifications, emit audit events.
	BeforeLogout func(ctx context.Context, sessionID uuid.UUID) error

	// AfterEmailVerified runs after a user's email is successfully verified.
	// It receives the raw JWT claims (access token) so the hook can extract
	// application-specific data embedded by ClaimsBuilder, such as a user's
	// slug. Errors returned by this hook are blocking: they abort the
	// verification flow and roll back the transaction.
	//
	// Use cases: auto-create profile via friendship service, send welcome email.
	AfterEmailVerified func(ctx context.Context, userID uuid.UUID, profileID uuid.UUID, claims map[string]any) error

	// AfterOAuthLogin runs after a successful OAuth authentication, regardless
	// of whether the user is new or returning. The isNewUser flag distinguishes
	// between first-time registration and existing account login. Errors
	// returned by this hook are non-blocking.
	//
	// Use cases: record referral for new OAuth users, sync profile data.
	AfterOAuthLogin func(ctx context.Context, userID uuid.UUID, provider string, isNewUser bool) error

	// ClaimsBuilder is called whenever a JWT access token is being issued
	// (during both login and token refresh). It receives the base claims
	// already set by the session service (sub, exp, iat, sid) and returns a
	// modified or augmented map to embed in the token. Errors returned by this
	// hook abort token issuance.
	//
	// This hook lives on AuthHooks for discoverability, but it is wired into
	// SessionService (not AuthService) because claims are built during token
	// creation. Extract ClaimsBuilder from the hooks value and pass it to
	// NewSessionService.
	//
	// Use cases: embed slug, tier, or other app-specific fields into the JWT.
	ClaimsBuilder func(ctx context.Context, userID uuid.UUID, baseClaims map[string]any) (map[string]any, error)
}

// CallBeforeLogout invokes BeforeLogout if it is non-nil. Errors are returned
// to the caller so the caller can decide whether to block (the hook contract
// says they should be logged, not returned).
func (h AuthHooks) CallBeforeLogout(ctx context.Context, sessionID uuid.UUID) error {
	if h.BeforeLogout == nil {
		return nil
	}
	return h.BeforeLogout(ctx, sessionID)
}

// CallAfterEmailVerified invokes AfterEmailVerified if it is non-nil.
func (h AuthHooks) CallAfterEmailVerified(ctx context.Context, userID, profileID uuid.UUID, claims map[string]any) error {
	if h.AfterEmailVerified == nil {
		return nil
	}
	return h.AfterEmailVerified(ctx, userID, profileID, claims)
}

// CallAfterOAuthLogin invokes AfterOAuthLogin if it is non-nil.
func (h AuthHooks) CallAfterOAuthLogin(ctx context.Context, userID uuid.UUID, provider string, isNewUser bool) error {
	if h.AfterOAuthLogin == nil {
		return nil
	}
	return h.AfterOAuthLogin(ctx, userID, provider, isNewUser)
}