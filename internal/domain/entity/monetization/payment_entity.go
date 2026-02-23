package monetization

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type Payment struct {
	crud.BaseEntity
	SubscriptionID        uuid.UUID
	Amount                decimal.Decimal
	Currency              string
	Gateway               string
	GatewayTransactionID  string
	GatewaySubscriptionID string
	Status                string
	FailureReason         sql.NullString
	StartsAt              sql.NullTime
	EndsAt                sql.NullTime
	GatewayEventID        sql.NullString
	PaidAt                sql.NullTime
}

func (Payment) TableName() string {
	return "subscription_payments"
}
