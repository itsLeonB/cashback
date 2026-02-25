package monetization

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/gorm"
)

type subscriptionRepository struct {
	crud.Repository[entity.Subscription]
}

func NewSubscriptionRepository(db *gorm.DB) *subscriptionRepository {
	return &subscriptionRepository{crud.NewRepository[entity.Subscription](db)}
}

func (sr *subscriptionRepository) UpdatePastDues(ctx context.Context) error {
	db, err := sr.GetGormInstance(ctx)
	if err != nil {
		return err
	}

	result := db.Model(&entity.Subscription{}).
		Where("current_period_end IS NOT NULL AND current_period_end < ? AND status != ?", time.Now(), entity.SubscriptionPastDuePayment).
		Update("status", entity.SubscriptionPastDuePayment)

	if err = result.Error; err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	logger.Infof("%d subscriptions is now having past due payments", result.RowsAffected)

	return nil
}

func (sr *subscriptionRepository) FindNearingDueDate(ctx context.Context) ([]entity.Subscription, error) {
	db, err := sr.GetGormInstance(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	in3Days := now.Add(3 * 24 * time.Hour)

	var subscriptions []entity.Subscription

	err = db.
		Preload("Profile").
		Where("current_period_end IS NOT NULL AND current_period_end > ? AND current_period_end < ? AND status = ?", now, in3Days, entity.SubscriptionActive).
		Find(&subscriptions).
		Error

	if err != nil {
		return nil, ungerr.Wrap(err, appconstant.ErrDataSelect)
	}

	return subscriptions, nil
}
