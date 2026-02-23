package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/go-crud"
)

type ExpenseBillRepository interface {
	crud.Repository[expenses.ExpenseBill]
	CountUploadedByDateRange(ctx context.Context, profileID uuid.UUID, start, end time.Time) (int, error)
}
