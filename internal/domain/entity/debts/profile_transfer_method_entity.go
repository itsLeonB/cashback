package debts

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
)

type ProfileTransferMethod struct {
	crud.BaseEntity
	ProfileID        uuid.UUID
	TransferMethodID uuid.UUID
	AccountName      string
	AccountNumber    string

	// Relationships
	Method TransferMethod `gorm:"foreignKey:TransferMethodID"`
}
