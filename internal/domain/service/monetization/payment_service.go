package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PaymentService interface {
	IsReady() error
	Create(ctx context.Context, req dto.NewPaymentRequest) (dto.PaymentResponse, error)
	HandleNotification(ctx context.Context, orderID string) error
}

func NewPaymentService(
	gateway payment.Gateway,
	transactor crud.Transactor,
	paymentRepo crud.Repository[entity.Payment],
) *paymentService {
	return &paymentService{
		gateway,
		transactor,
		paymentRepo,
	}
}

type paymentService struct {
	gateway     payment.Gateway
	transactor  crud.Transactor
	paymentRepo crud.Repository[entity.Payment]
}

func (ps *paymentService) IsReady() error {
	if ps.gateway == nil {
		return ungerr.Unknown("payment gateway is unitialized")
	}
	return nil
}

func (ps *paymentService) Create(ctx context.Context, req dto.NewPaymentRequest) (dto.PaymentResponse, error) {
	var resp dto.PaymentResponse
	var err error
	err = ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		newPayment := entity.Payment{
			SubscriptionID: req.SubscriptionID,
			Amount:         req.Amount,
			Currency:       req.Currency,
			Status:         entity.PendingPayment,
			Gateway:        ps.gateway.Provider(),
		}

		pendingPayment, err := ps.paymentRepo.Insert(ctx, newPayment)
		if err != nil {
			return err
		}

		requestedPayment, err := ps.gateway.CreateTransaction(ctx, pendingPayment)
		if err != nil {
			return err
		}

		requestedPayment, err = ps.paymentRepo.Update(ctx, requestedPayment)
		if err != nil {
			return err
		}

		resp = mapper.PaymentToResponse(requestedPayment)
		return nil
	})
	return resp, err
}

func (ps *paymentService) HandleNotification(ctx context.Context, orderID string) error {
	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		id, err := ezutil.Parse[uuid.UUID](orderID)
		if err != nil {
			return err
		}

		spec := crud.Specification[entity.Payment]{}
		spec.Model.ID = id
		spec.Model.Gateway = ps.gateway.Provider()
		spec.ForUpdate = true
		payment, err := ps.paymentRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if payment.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("payment with ID %s is not found", id))
		}

		newStatus, err := ps.gateway.CheckStatus(ctx, orderID)
		if err != nil {
			logger.Error(err)
		}
		if newStatus == "" {
			return nil
		}

		payment.Status = newStatus

		if newStatus == entity.ErrorPayment {
			payment.FailureReason = sql.NullString{
				String: err.Error(),
				Valid:  true,
			}
		}

		if newStatus == entity.PaidPayment {
			payment.PaidAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
		}

		_, err = ps.paymentRepo.Update(ctx, payment)
		return err
	})
}
