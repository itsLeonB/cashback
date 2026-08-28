package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type friendshipRequestRepositoryGorm struct {
	crud.Repository[users.FriendshipRequest]
}

func NewFriendshipRequestRepository(db *gorm.DB) *friendshipRequestRepositoryGorm {
	return &friendshipRequestRepositoryGorm{
		crud.NewRepository[users.FriendshipRequest](db),
	}
}

// RepointProfile repoints every friendship request referencing anonProfileID (as sender or
// recipient) onto realProfileID. No unique constraint on these columns, so this is a plain update.
func (fr *friendshipRequestRepositoryGorm) RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "FriendshipRequestRepository.RepointProfile")
	defer span.End()

	db, err := fr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	if err := db.Model(&users.FriendshipRequest{}).
		Where("sender_profile_id = ?", anonProfileID).
		Update("sender_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}
	if err := db.Model(&users.FriendshipRequest{}).
		Where("recipient_profile_id = ?", anonProfileID).
		Update("recipient_profile_id", realProfileID).Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
