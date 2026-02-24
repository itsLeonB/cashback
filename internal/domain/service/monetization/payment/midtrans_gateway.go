package payment

import (
	"context"
	"database/sql"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransGateway struct {
	snapClient *snap.Client
	coreClient *coreapi.Client
}

func newMidtransGateway(cfg config.Payment) (*midtransGateway, error) {
	snapClient := &snap.Client{}
	coreClient := &coreapi.Client{}

	env, err := loadMidtransEnv(cfg.Env)
	if err != nil {
		return nil, err
	}

	snapClient.New(cfg.ServerKey, env)
	coreClient.New(cfg.ServerKey, env)

	return &midtransGateway{snapClient, coreClient}, nil
}

func loadMidtransEnv(envCfg string) (midtrans.EnvironmentType, error) {
	switch envCfg {
	case "sandbox":
		return midtrans.Sandbox, nil
	case "production":
		return midtrans.Production, nil
	default:
		return 0, ungerr.Unknownf("unsupported payment gateway env: %s", envCfg)
	}
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

	resp, err := mg.snapClient.CreateTransaction(req)
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

func (mg *midtransGateway) CheckStatus(ctx context.Context, orderID string) (entity.PaymentStatus, error) {
	trxStatusResp, err := mg.coreClient.CheckTransaction(orderID)
	if err != nil {
		return entity.ErrorPayment, ungerr.Wrapf(err, "error checking transaction status of ID: %s", orderID)
	}

	switch trxStatusResp.TransactionStatus {
	case "capture":
		switch trxStatusResp.FraudStatus {
		case "challenge":
			logger.Warn("received fraud challenge, please check midtrans dashboard")
			return entity.ProcessingPayment, nil
		case "accept":
			return entity.PaidPayment, nil
		default:
			return entity.ErrorPayment, ungerr.Unknownf("unhandled fraud status: %s", trxStatusResp.FraudStatus)
		}
	case "settlement":
		return entity.PaidPayment, nil
	case "deny":
		return "", nil
	case "cancel", "expire":
		return entity.CanceledPayment, nil
	case "pending":
		return entity.PendingPayment, nil
	default:
		return entity.ErrorPayment, ungerr.Unknownf("unhandled transaction status: %s", trxStatusResp.TransactionStatus)
	}
}
