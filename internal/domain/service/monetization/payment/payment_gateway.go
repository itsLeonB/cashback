package payment

import (
	"time"

	"github.com/google/uuid"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
)

type Gateway interface {
	CreateCheckoutSession(params CheckoutParams) (CheckoutResult, error)
	HandleWebhook(payload []byte, signature string) (*WebhookEvent, error)
	CreatePortalSession(customerID string, returnURL string) (string, error)
}

type CheckoutParams struct {
	Payment      entity.Payment
	PlanVersion  entity.PlanVersion
	CustomerID   string
	ProfileEmail string
	ProfileID    uuid.UUID
	SuccessURL   string
	CancelURL    string
}

type CheckoutResult struct {
	CheckoutURL       string
	GatewayCustomerID string
	GatewaySessionID  string
}

type WebhookEvent struct {
	Type           string // "payment_success", "payment_failed", "subscription_canceled"
	GatewayEventID string
	SubscriptionID uuid.UUID
	GatewaySubID   string
	PeriodStart    time.Time
	PeriodEnd      time.Time
}
