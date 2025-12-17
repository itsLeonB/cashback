package expensebill

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/domain"
)

type ExpenseBill struct {
	CreatorProfileID uuid.UUID `validate:"required"`
	PayerProfileID   uuid.UUID `validate:"required"`
	GroupExpenseID   uuid.UUID
	ObjectKey        string `validate:"required"`
	domain.AuditMetadata
}
