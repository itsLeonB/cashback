package monetization

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/shopspring/decimal"
)

type NewPaymentRequest struct {
	SubscriptionID uuid.UUID
	Currency       string
	Amount         decimal.Decimal
}

type PaymentResponse struct {
	dto.BaseDTO
	SubscriptionID        uuid.UUID       `json:"subscriptionId"`
	Amount                decimal.Decimal `json:"amount"`
	Currency              string          `json:"currency"`
	Gateway               string          `json:"gateway"`
	GatewayTransactionID  string          `json:"gatewayTransactionId,omitzero"`
	GatewaySubscriptionID string          `json:"gatewaySubscriptionId,omitzero"`
	Status                string          `json:"status"`
	FailureReason         string          `json:"failureReason,omitzero"`
	StartsAt              time.Time       `json:"startsAt,omitzero"`
	EndsAt                time.Time       `json:"endsAt,omitzero"`
	GatewayEventID        string          `json:"gatewayEventId,omitzero"`
	PaidAt                time.Time       `json:"paidAt,omitzero"`
	ExpiredAt             time.Time       `json:"expiredAt,omitzero"`
}

type MidtransNotificationPayload struct {
	OrderID       string `json:"order_id" binding:"required"`
	StatusCode    string `json:"status_code" binding:"required"`
	GrossAmount   string `json:"gross_amount" binding:"required"`
	SignatureKey  string `json:"signature_key" binding:"required"`
	StatusMessage string `json:"status_message"`
}

type UpdatePaymentRequest struct {
	ID       uuid.UUID       `json:"-"`
	Status   string          `json:"status" binding:"required,oneof=pending processing paid canceled error expired"`
	Amount   decimal.Decimal `json:"amount" binding:"required"`
	Currency string          `json:"currency" binding:"required"`
	StartsAt time.Time       `json:"startsAt"`
	EndsAt   time.Time       `json:"endsAt"`
	PaidAt   time.Time       `json:"paidAt"`
}
