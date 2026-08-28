package users

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

// FriendshipBalance caches the net debt balance for a friendship pair, per currency.
// NetBalance is relative to the owning Friendship's ProfileID1: positive means
// ProfileID1 is the net lender. Recomputed from debt_transactions on every write, never
// incremented, so it self-corrects regardless of call order.
type FriendshipBalance struct {
	crud.BaseEntity
	FriendshipID uuid.UUID
	Currency     string
	NetBalance   decimal.Decimal

	Friendship Friendship `gorm:"foreignKey:FriendshipID"`
}
