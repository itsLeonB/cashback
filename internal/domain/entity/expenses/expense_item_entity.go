package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ExpenseItem struct {
	crud.BaseEntity
	GroupExpenseID uuid.UUID
	Name           string
	Amount         decimal.Decimal
	Quantity       int
	Participants   []ItemParticipant `gorm:"foreignKey:ExpenseItemID"`
}

func (ei ExpenseItem) TableName() string {
	return "group_expense_items"
}

func (ei ExpenseItem) SimpleName() string {
	return "expense item"
}

func (ei ExpenseItem) TotalAmount() decimal.Decimal {
	return ei.Amount.Mul(decimal.NewFromInt(int64(ei.Quantity)))
}

func (ei ExpenseItem) ProfileIDs() []uuid.UUID {
	return ezutil.MapSlice(ei.Participants, func(ip ItemParticipant) uuid.UUID { return ip.ProfileID })
}
