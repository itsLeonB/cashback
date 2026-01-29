package repository

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
)

type PushSubscriptionRepository interface {
	crud.Repository[entity.PushSubscription]
	Upsert(ctx context.Context, subscription entity.PushSubscription) error
}
