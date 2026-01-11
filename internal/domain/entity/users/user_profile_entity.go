package users

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type UserProfile struct {
	crud.BaseEntity
	UserID uuid.NullUUID
	Name   string
	Avatar string

	// Relationships
	RelatedRealProfile  RelatedProfile   `gorm:"foreignKey:AnonProfileID"`
	RelatedAnonProfiles []RelatedProfile `gorm:"foreignKey:RealProfileID"`
}

func (up UserProfile) IsReal() bool {
	return up.UserID.Valid
}
