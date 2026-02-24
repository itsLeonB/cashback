package payment

import (
	"context"
	"database/sql"

	"github.com/itsLeonB/cashback/internal/core/config"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransGateway struct {
	client *snap.Client
}

func newMidtransGateway(cfg config.Payment) (*midtransGateway, error) {
	client := &snap.Client{}

	var env midtrans.EnvironmentType
	switch cfg.Env {
	case "sandbox":
		env = midtrans.Sandbox
	case "production":
		env = midtrans.Production
	default:
		return nil, ungerr.Unknownf("unsupported payment gateway env: %s", cfg.Env)
	}

	client.New(cfg.ServerKey, env)

	return &midtransGateway{client}, nil
}

func (mg *midtransGateway) Provider() string {
	return "midtrans"
}

func (mg *midtransGateway) CreateTransaction(ctx context.Context, payment entity.Payment) (entity.Payment, error) {
	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  payment.ID.String(),
			GrossAmt: payment.Amount.IntPart(),
		},
	}

	resp, err := mg.client.CreateTransaction(req)
	if err != nil {
		return entity.Payment{}, ungerr.Wrap(err, "error creating midtrans transaction")
	}

	if resp.StatusCode[0] != '2' {
		return entity.Payment{}, ungerr.Unknownf("non-success status code: %s", resp.StatusCode)
	}

	payment.GatewayTransactionID = sql.NullString{
		String: resp.Token,
		Valid:  true,
	}

	return payment, nil
}
