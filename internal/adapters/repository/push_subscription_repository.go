package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pushSubscriptionRepositoryGorm struct {
	crud.Repository[entity.PushSubscription]
}

func NewPushSubscriptionRepository(db *gorm.DB) *pushSubscriptionRepositoryGorm {
	return &pushSubscriptionRepositoryGorm{
		crud.NewRepository[entity.PushSubscription](db),
	}
}

func (r *pushSubscriptionRepositoryGorm) Upsert(ctx context.Context, subscription entity.PushSubscription) error {
	ctx, span := otel.Tracer.Start(ctx, "PushSubscriptionRepository.Upsert")
	defer span.End()

	db, err := r.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err = db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "endpoint"}},
		DoUpdates: clause.AssignmentColumns([]string{"keys", "user_agent"}),
	}).Create(&subscription).Error; err != nil {
		return ungerr.Wrap(err, "failed to upsert push subscription")
	}

	return nil
}

// RepointProfile repoints push subscriptions referencing anonProfileID onto realProfileID. If
// realProfileID already has a subscription for the same endpoint, the anonymous row is dropped
// instead of violating push_subscriptions_profile_endpoint_unique_idx. Anonymous profiles never
// log in, so in practice this has nothing to do.
func (r *pushSubscriptionRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "PushSubscriptionRepository.RepointProfile")
	defer span.End()

	db, err := r.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	var anonRows []entity.PushSubscription
	if err := db.Where("profile_id = ?", anonProfileID).Find(&anonRows).Error; err != nil {
		return ungerr.Wrap(err, "failed to select push subscriptions")
	}
	if len(anonRows) == 0 {
		return nil
	}

	moved := make([]entity.PushSubscription, len(anonRows))
	anonRowIDs := make([]uuid.UUID, len(anonRows))
	for i, s := range anonRows {
		moved[i] = entity.PushSubscription{
			ProfileID: realProfileID,
			SessionID: s.SessionID,
			Endpoint:  s.Endpoint,
			Keys:      s.Keys,
			UserAgent: s.UserAgent,
		}
		anonRowIDs[i] = s.ID
	}

	// Anon profiles never log in, so in practice there's nothing to merge here; if the real
	// profile already has a subscription for the same endpoint, keep it and drop the anon
	// duplicate below.
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "profile_id"}, {Name: "endpoint"}},
		DoNothing: true,
	}).Create(&moved).Error; err != nil {
		return ungerr.Wrap(err, "failed to repoint push subscriptions")
	}

	// Restricted to the exact rows read above so a row inserted concurrently after the
	// SELECT - and never merged into realProfileID - survives instead of being silently
	// dropped.
	if err := db.Where("id IN ?", anonRowIDs).Delete(&entity.PushSubscription{}).Error; err != nil {
		return ungerr.Wrap(err, "error deleting merged push subscriptions")
	}

	return nil
}
