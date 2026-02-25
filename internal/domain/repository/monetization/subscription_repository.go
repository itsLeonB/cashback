package monetization

import (
	"context"

	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/go-crud"
)

type SubscriptionRepository interface {
	crud.Repository[entity.Subscription]
	UpdatePastDues(ctx context.Context) error
}
