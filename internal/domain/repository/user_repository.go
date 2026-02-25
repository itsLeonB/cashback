package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
)

type UserRepository interface {
	crud.Repository[users.User]
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.User, error)
}
