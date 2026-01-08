package expenses

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type FeeParticipant struct {
	crud.BaseEntity
	OtherFeeID  uuid.UUID
	ProfileID   uuid.UUID
	ShareAmount decimal.Decimal

	// Relationships
	Profile users.UserProfile `gorm:"foreignKey:ProfileID"`
}

func (fp FeeParticipant) TableName() string {
	return "group_expense_other_fee_participants"
}
