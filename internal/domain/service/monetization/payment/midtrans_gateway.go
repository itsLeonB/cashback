package payment

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"fmt"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/ungerr"
	"github.com/midtrans/midtrans-go"
	"github.com/midtrans/midtrans-go/coreapi"
	"github.com/midtrans/midtrans-go/snap"
)

type midtransGateway struct {
	snapClient *snap.Client
	coreClient *coreapi.Client
	serverKey  string
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

	return &midtransGateway{snapClient, coreClient, cfg.ServerKey}, nil
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

	token, err := mg.snapClient.CreateTransactionToken(req)
	if err != nil {
		return entity.Payment{}, ungerr.Wrap(err, "error creating midtrans transaction")
	}

	payment.GatewayTransactionID = sql.NullString{
		String: token,
		Valid:  true,
	}

	return payment, nil
}

func (mg *midtransGateway) CheckStatus(ctx context.Context, req dto.MidtransNotificationPayload) (entity.PaymentStatus, error) {
	if err := mg.validate(req); err != nil {
		return entity.ErrorPayment, err
	}

	trxStatusResp, err := mg.coreClient.CheckTransaction(req.OrderID)
	if err != nil {
		return entity.ErrorPayment, ungerr.Wrapf(err, "error checking transaction status of ID: %s", req.OrderID)
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

func (mg *midtransGateway) validate(req dto.MidtransNotificationPayload) error {
	checkKey := req.OrderID + req.StatusCode + req.GrossAmount + mg.serverKey
	constructedKey := sha512.Sum512([]byte(checkKey))

	if fmt.Sprint(constructedKey) == req.SignatureKey {
		return nil
	}

	return ungerr.Unknown("signature key cannot be validated")
}
