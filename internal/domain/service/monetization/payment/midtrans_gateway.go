package payment

import (
	"context"
	"crypto/sha512"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
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

	env := midtrans.Production
	if strings.Contains(cfg.BaseUrl, "sandbox") {
		env = midtrans.Sandbox
	}

	snapClient.New(cfg.ServerKey, env)
	coreClient.New(cfg.ServerKey, env)

	return &midtransGateway{snapClient, coreClient, cfg.ServerKey}, nil
}


func (mg *midtransGateway) Provider() string {
	return "midtrans"
}

func (mg *midtransGateway) CreateTransaction(ctx context.Context, payment entity.Payment) (entity.Payment, error) {
	ctx, span := otel.Tracer.Start(ctx, "midtransGateway.CreateTransaction")
	defer span.End()

	req := &snap.Request{
		TransactionDetails: midtrans.TransactionDetails{
			OrderID:  payment.ID.String(),
			GrossAmt: payment.Amount.IntPart(),
		},
	}

	snapClient := *mg.snapClient
	snapClient.Options = &midtrans.ConfigOptions{}
	snapClient.Options.SetContext(ctx)
	token, err := snapClient.CreateTransactionToken(req)
	if err != nil {
		return entity.Payment{}, ungerr.Wrap(err, "error creating midtrans transaction")
	}

	payment.GatewayTransactionID = sql.NullString{
		String: token,
		Valid:  true,
	}

	return payment, nil
}

func (mg *midtransGateway) ValidateAndCheckStatus(ctx context.Context, payload NotificationPayload) (entity.PaymentStatus, error) {
	ctx, span := otel.Tracer.Start(ctx, "midtransGateway.ValidateAndCheckStatus")
	defer span.End()

	// Validate signature
	statusCode := payload.Extra["status_code"]
	grossAmount := payload.Extra["gross_amount"]
	signatureKey := payload.Signature

	checkKey := payload.OrderID + statusCode + grossAmount + mg.serverKey
	constructedKey := sha512.Sum512([]byte(checkKey))
	if fmt.Sprintf("%x", constructedKey) != signatureKey {
		return entity.ErrorPayment, ungerr.Unknown("signature key cannot be validated")
	}

	coreClient := *mg.coreClient
	coreClient.Options = &midtrans.ConfigOptions{}
	coreClient.Options.SetContext(ctx)
	trxStatusResp, err := coreClient.CheckTransaction(payload.OrderID)
	if err != nil {
		return entity.ErrorPayment, ungerr.Wrapf(err, "error checking transaction status of ID: %s", payload.OrderID)
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
		statusMessage := payload.Extra["status_message"]
		var e error
		if statusMessage == "" {
			e = errors.New("unknown")
		} else {
			e = errors.New(statusMessage)
		}
		return entity.ErrorPayment, e
	case "cancel", "expire":
		return entity.CanceledPayment, nil
	case "pending":
		return entity.PendingPayment, nil
	default:
		return entity.ErrorPayment, ungerr.Unknownf("unhandled transaction status: %s", trxStatusResp.TransactionStatus)
	}
}
