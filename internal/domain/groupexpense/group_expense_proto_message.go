package groupexpense

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/shopspring/decimal"
)

type CreateDraftRequest struct {
	CreatorProfileID uuid.UUID       `validate:"required"`
	PayerProfileID   uuid.UUID       `validate:"required"`
	TotalAmount      decimal.Decimal `validate:"required"`
	Subtotal         decimal.Decimal `validate:"required"`
	Description      string
	Items            []expenseitem.ExpenseItemData `validate:"required,min=1,dive"`
	OtherFees        []otherfee.OtherFeeData       `validate:"dive"`
}

type ConfirmDraftRequest struct {
	ID        uuid.UUID `validate:"required"`
	ProfileID uuid.UUID `validate:"required"`
	DryRun    bool
}
