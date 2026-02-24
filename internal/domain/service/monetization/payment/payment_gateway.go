package payment

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/config"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
)

type Gateway interface {
	Provider() string
	CreateTransaction(ctx context.Context, payment entity.Payment) (entity.Payment, error)
}

func NewGateway(cfg config.Payment) (Gateway, error) {
	switch cfg.Gateway {
	case "midtrans":
		return newMidtransGateway(cfg)
	default:
		return nil, ungerr.Unknownf("unsupported payment gateway: %s", cfg.Gateway)
	}
}
