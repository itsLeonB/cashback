package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ExpenseStatus string

const (
	DraftExpense     ExpenseStatus = "DRAFT"
	ReadyExpense     ExpenseStatus = "READY"
	ConfirmedExpense ExpenseStatus = "CONFIRMED"
)

type GroupExpense struct {
	crud.BaseEntity
	PayerProfileID   uuid.UUID
	TotalAmount      decimal.Decimal
	ItemsTotal       decimal.Decimal
	FeesTotal        decimal.Decimal
	Description      string
	Status           ExpenseStatus
	CreatorProfileID uuid.UUID

	// Relationships
	Payer        users.UserProfile    `gorm:"foreignKey:PayerProfileID"`
	Creator      users.UserProfile    `gorm:"foreignKey:CreatorProfileID"`
	Items        []ExpenseItem        `gorm:"foreignKey:GroupExpenseID"`
	OtherFees    []OtherFee           `gorm:"foreignKey:GroupExpenseID"`
	Participants []ExpenseParticipant `gorm:"foreignKey:GroupExpenseID"`
	Bill         ExpenseBill          `gorm:"foreignKey:GroupExpenseID"`
}
