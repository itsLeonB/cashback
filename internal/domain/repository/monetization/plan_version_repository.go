package monetization

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/go-crud"
)

type PlanVersionRepository interface {
	crud.Repository[monetization.PlanVersion]
	SetAsDefault(ctx context.Context, id uuid.UUID) error
}
