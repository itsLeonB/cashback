package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ExpenseParticipant struct {
	crud.BaseEntity
	GroupExpenseID       uuid.UUID
	ParticipantProfileID uuid.UUID
	ShareAmount          decimal.Decimal
}

func (ep ExpenseParticipant) TableName() string {
	return "group_expense_participants"
}
