package debts

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type DebtTransactionType string
type DebtTransactionAction string

const (
	Lend  DebtTransactionType = "LEND"
	Repay DebtTransactionType = "REPAY"

	LendAction    DebtTransactionAction = "LEND"
	BorrowAction  DebtTransactionAction = "BORROW"
	ReceiveAction DebtTransactionAction = "RECEIVE"
	ReturnAction  DebtTransactionAction = "RETURN"

	GroupExpenseTransferMethod = "GROUP_EXPENSE"
)

type DebtTransaction struct {
	crud.BaseEntity
	LenderProfileID   uuid.UUID
	BorrowerProfileID uuid.UUID
	Type              DebtTransactionType
	Action            DebtTransactionAction
	Amount            decimal.Decimal
	TransferMethodID  uuid.UUID
	Description       string
	TransferMethod    TransferMethod
}
