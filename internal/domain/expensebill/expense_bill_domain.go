package expensebill

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain"
)

type ExpenseBill struct {
	CreatorProfileID uuid.UUID `validate:"required"`
	PayerProfileID   uuid.UUID
	GroupExpenseID   uuid.UUID
	ObjectKey        string `validate:"required"`
	Status           appconstant.BillStatus
	domain.AuditMetadata
}
