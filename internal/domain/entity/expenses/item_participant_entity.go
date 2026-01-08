package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type ItemParticipant struct {
	crud.BaseEntity
	ExpenseItemID uuid.UUID
	ProfileID     uuid.UUID
	Share         decimal.Decimal

	// Relationships
	Profile users.UserProfile `gorm:"foreignKey:ProfileID"`
}

func (ip ItemParticipant) TableName() string {
	return "group_expense_item_participants"
}
