package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ItemParticipant struct {
	crud.BaseEntity
	ExpenseItemID uuid.UUID
	ProfileID     uuid.UUID
	Share         decimal.Decimal
}

func (ip ItemParticipant) TableName() string {
	return "group_expense_item_participants"
}
