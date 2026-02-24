package monetization

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/shopspring/decimal"
)

type PaymentResponse struct {
	dto.BaseDTO
	SubscriptionID        uuid.UUID       `json:"subscriptionId"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Gateway               string          `json:"gateway"`
	GatewayTransactionID  string          `json:"gatewayTransactionId"`
	GatewaySubscriptionID string          `json:"gatewaySubscriptionId,omitzero"`
	Status                string          `json:"status"`
	FailureReason         string          `json:"failureReason,omitzero"`
	StartsAt              time.Time       `json:"startsAt,omitzero"`
	EndsAt                time.Time       `json:"endsAt,omitzero"`
	GatewayEventID        string          `json:"gatewayEventId,omitzero"`
	PaidAt                time.Time       `json:"paidAt"`
}
