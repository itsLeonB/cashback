package monetization

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/appconstant"
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

	err = db.Model(&entity.Subscription{}).
		Where("current_period_end < ? AND status != ?", time.Now(), entity.SubscriptionPastDuePayment).
		Update("status", entity.SubscriptionPastDuePayment).
		Error

	if err != nil {
		return ungerr.Wrap(err, appconstant.ErrDataUpdate)
	}

	return nil
}
