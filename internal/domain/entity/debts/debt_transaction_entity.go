package debts

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

const (
	GroupExpenseTransferMethod = "GROUP_EXPENSE"
)

type DebtTransaction struct {
	crud.BaseEntity
	LenderProfileID   uuid.UUID
	BorrowerProfileID uuid.UUID
	Amount            decimal.Decimal
	TransferMethodID  uuid.UUID
	Description       string
	GroupExpenseID    uuid.NullUUID

	// Relationships
	TransferMethod TransferMethod
}
