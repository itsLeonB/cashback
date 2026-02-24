package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PaymentService interface {
	IsReady() error
	Create(ctx context.Context, req dto.NewPaymentRequest) (dto.PaymentResponse, error)
	HandleNotification(ctx context.Context, req dto.MidtransNotificationPayload) error
}

func NewPaymentService(
	gateway payment.Gateway,
	transactor crud.Transactor,
	paymentRepo crud.Repository[entity.Payment],
	taskQueue queue.TaskQueue,
) *paymentService {
	return &paymentService{
		gateway,
		transactor,
		paymentRepo,
		taskQueue,
	}
}

type paymentService struct {
	gateway     payment.Gateway
	transactor  crud.Transactor
	paymentRepo crud.Repository[entity.Payment]
	taskQueue   queue.TaskQueue
}

func (ps *paymentService) IsReady() error {
	if ps.gateway == nil {
		return ungerr.Unknown("payment gateway is uninitialized")
	}
	return nil
}

func (ps *paymentService) Create(ctx context.Context, req dto.NewPaymentRequest) (dto.PaymentResponse, error) {
	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
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

func (ps *paymentService) HandleNotification(ctx context.Context, req dto.MidtransNotificationPayload) error {
	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		id, err := ezutil.Parse[uuid.UUID](req.OrderID)
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
		if payment.Status == entity.PaidPayment {
			return nil
		}

		newStatus, err := ps.gateway.CheckStatus(ctx, req)
		if err != nil {
			logger.Error(err)
		}
		if newStatus == "" || newStatus == payment.Status {
			return nil
		}

		if err = ps.updatePaymentStatus(ctx, payment, newStatus, err); err != nil {
			return err
		}

		return ps.queueSubscriptionTransitioned(ctx, newStatus, payment.SubscriptionID)
	})
}

func (ps *paymentService) updatePaymentStatus(ctx context.Context, payment entity.Payment, newStatus entity.PaymentStatus, statusErr error) error {
	payment.Status = newStatus

	if newStatus == entity.ErrorPayment && statusErr != nil {
		payment.FailureReason = sql.NullString{
			String: statusErr.Error(),
			Valid:  true,
		}
	}

	if newStatus == entity.PaidPayment {
		payment.PaidAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	_, err := ps.paymentRepo.Update(ctx, payment)
	return err
}

func (ps *paymentService) queueSubscriptionTransitioned(ctx context.Context, newStatus entity.PaymentStatus, subID uuid.UUID) error {
	var subStatus entity.SubscriptionStatus
	switch newStatus {
	case entity.PaidPayment:
		subStatus = entity.SubscriptionActive
	case entity.CanceledPayment:
		subStatus = entity.SubscriptionCanceled
	case entity.ErrorPayment:
		subStatus = entity.SubscriptionCanceled
	case entity.PendingPayment:
		subStatus = entity.SubscriptionIncompletePayment
	case entity.ProcessingPayment:
		subStatus = entity.SubscriptionIncompletePayment
	default:
		return ungerr.Unknownf("unhandled payment status: %s", newStatus)
	}

	msg := message.SubscriptionStatusTransitioned{
		ID:     subID,
		Status: subStatus,
	}

	return ps.taskQueue.Enqueue(ctx, msg)
}
