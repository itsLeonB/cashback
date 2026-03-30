package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ExpenseParticipant struct {
	crud.BaseEntity
	GroupExpenseID       uuid.UUID
	ParticipantProfileID uuid.UUID
	ProxyProfileID       uuid.NullUUID
	ShareAmount          decimal.Decimal

	// Relationships
	ParticipantProfile users.UserProfile `gorm:"foreignKey:ParticipantProfileID"`
	ProxyProfile       users.UserProfile `gorm:"foreignKey:ProxyProfileID"`
}

func (ep ExpenseParticipant) TableName() string {
	return "group_expense_participants"
}
