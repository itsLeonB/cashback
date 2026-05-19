package users

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/go-crud"
)

type UserProfile struct {
	crud.BaseEntity
	UserID           uuid.NullUUID
	Name             string
	Avatar           string
	HomeCurrency     string
	OnboardedAt      sql.NullTime
	StripeCustomerID sql.NullString

	// Relationships
	RelatedRealProfile  RelatedProfile                `gorm:"foreignKey:AnonProfileID"`
	RelatedAnonProfiles []RelatedProfile              `gorm:"foreignKey:RealProfileID"`
	TransferMethods     []debts.ProfileTransferMethod `gorm:"foreignKey:ProfileID"`
}

func (up UserProfile) IsReal() bool {
	return up.UserID.Valid
}
