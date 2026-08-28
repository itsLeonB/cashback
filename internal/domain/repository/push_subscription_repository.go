package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
)

type PushSubscriptionRepository interface {
	crud.Repository[entity.PushSubscription]
	Upsert(ctx context.Context, subscription entity.PushSubscription) error
	// RepointProfile repoints push subscriptions referencing anonProfileID onto realProfileID.
	// If realProfileID already has a subscription for the same endpoint, the anonymous row is
	// dropped instead of violating the unique(profile_id, endpoint) constraint. Anonymous
	// profiles never log in, so in practice this is a no-op.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}
