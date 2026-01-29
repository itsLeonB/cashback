package repository

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pushSubscriptionRepositoryGorm struct {
	crud.Repository[entity.PushSubscription]
	db *gorm.DB
}

func NewPushSubscriptionRepository(db *gorm.DB) *pushSubscriptionRepositoryGorm {
	return &pushSubscriptionRepositoryGorm{db: db}
}

func (r *pushSubscriptionRepositoryGorm) Upsert(ctx context.Context, subscription entity.PushSubscription) error {
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
