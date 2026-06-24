package payment

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/config"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
)

// NotificationPayload is a gateway-agnostic representation of an incoming payment notification.
type NotificationPayload struct {
	OrderID   string // payment UUID
	RawStatus string // gateway-native status string
	Signature string // for validation
	Extra     map[string]string
}

type Gateway interface {
	Provider() string
	CreateTransaction(ctx context.Context, payment entity.Payment) (entity.Payment, error)
	ValidateAndCheckStatus(ctx context.Context, payload NotificationPayload) (entity.PaymentStatus, error)
}

func NewGateway(cfg config.Payment) (Gateway, error) {
	switch cfg.Gateway {
	case "ipaymu":
		return newIPaymuGateway(cfg)
	case "midtrans":
		return newMidtransGateway(cfg)
	default:
		return nil, ungerr.Unknownf("unsupported payment gateway: %s", cfg.Gateway)
	}
}
