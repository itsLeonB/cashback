package monetization

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type PaymentStatus string

const (
	PendingPayment    = "pending"
	ProcessingPayment = "processing"
	PaidPayment       = "paid"
	CanceledPayment   = "canceled"
	ErrorPayment      = "error"
	ExpiredPayment    = "expired"
)

type Payment struct {
	crud.BaseEntity
	SubscriptionID        uuid.UUID
	Amount                decimal.Decimal
	Currency              string
	Gateway               string
	GatewayTransactionID  sql.NullString
	GatewaySubscriptionID sql.NullString
	Status                PaymentStatus
	FailureReason         sql.NullString
	StartsAt              sql.NullTime
	EndsAt                sql.NullTime
	GatewayEventID        sql.NullString
	PaidAt                sql.NullTime
	ExpiredAt             sql.NullTime
}

func (Payment) TableName() string {
	return "subscription_payments"
}
