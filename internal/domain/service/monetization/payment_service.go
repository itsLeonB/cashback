package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
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
	NewPurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error)
	HandleNotification(ctx context.Context, req dto.MidtransNotificationPayload) error
	MakePayment(ctx context.Context, subscriptionID uuid.UUID) (dto.PaymentResponse, error)
}

func NewPaymentService(
	gateway payment.Gateway,
	transactor crud.Transactor,
	paymentRepo crud.Repository[entity.Payment],
	taskQueue queue.TaskQueue,
	subscriptionSvc SubscriptionService,
) *paymentService {
	return &paymentService{
		gateway,
		transactor,
		paymentRepo,
		taskQueue,
		subscriptionSvc,
	}
}

type paymentService struct {
	gateway         payment.Gateway
	transactor      crud.Transactor
	paymentRepo     crud.Repository[entity.Payment]
	taskQueue       queue.TaskQueue
	subscriptionSvc SubscriptionService
}

func (ps *paymentService) IsReady() error {
	if !config.Global.SubscriptionPurchaseEnabled {
		return ungerr.ForbiddenError("feature is disabled")
	}
	if ps.gateway == nil {
		return ungerr.Unknown("payment gateway is uninitialized")
	}
	return nil
}

func (ps *paymentService) NewPurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error) {
	if err := ps.IsReady(); err != nil {
		return dto.PaymentResponse{}, err
	}

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		paymentRequest, err := ps.subscriptionSvc.CreateNew(ctx, req)
		if err != nil {
			return err
		}

		resp, err = ps.create(ctx, paymentRequest)
		return err
	})
	return resp, err
}

func (ps *paymentService) create(ctx context.Context, req dto.NewPaymentRequest) (dto.PaymentResponse, error) {
	newPayment := entity.Payment{
		SubscriptionID: req.SubscriptionID,
		Amount:         req.Amount,
		Currency:       req.Currency,
		Status:         entity.PendingPayment,
		Gateway:        ps.gateway.Provider(),
	}

	pendingPayment, err := ps.paymentRepo.Insert(ctx, newPayment)
	if err != nil {
		return dto.PaymentResponse{}, err
	}

	requestedPayment, err := ps.gateway.CreateTransaction(ctx, pendingPayment)
	if err != nil {
		return dto.PaymentResponse{}, err
	}

	requestedPayment, err = ps.paymentRepo.Update(ctx, requestedPayment)
	if err != nil {
		return dto.PaymentResponse{}, err
	}

	return mapper.PaymentToResponse(requestedPayment), nil
}

func (ps *paymentService) HandleNotification(ctx context.Context, req dto.MidtransNotificationPayload) error {
	if err := ps.IsReady(); err != nil {
		return err
	}

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

		subs, err := ps.subscriptionSvc.GetByID(ctx, payment.SubscriptionID, false)
		if err != nil {
			return err
		}

		startsAt := time.Now()
		endsAt := startsAt
		switch subs.PlanVersion.BillingInterval {
		case entity.MonthlyInterval:
			endsAt = endsAt.AddDate(0, 1, 0)
		case entity.YearlyInterval:
			endsAt = endsAt.AddDate(1, 0, 0)
		}

		if err = ps.updatePaymentStatus(ctx, payment, newStatus, err, startsAt, endsAt); err != nil {
			return err
		}

		return ps.queueSubscriptionTransitioned(ctx, newStatus, payment.SubscriptionID)
	})
}

func (ps *paymentService) MakePayment(ctx context.Context, subscriptionID uuid.UUID) (dto.PaymentResponse, error) {
	if err := ps.IsReady(); err != nil {
		return dto.PaymentResponse{}, err
	}

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ps.subscriptionSvc.GetByID(ctx, subscriptionID, true)
		if err != nil {
			return err
		}

		req := dto.NewPaymentRequest{
			SubscriptionID: subscriptionID,
			Currency:       subscription.PlanVersion.PriceCurrency,
			Amount:         subscription.PlanVersion.PriceAmount,
		}

		resp, err = ps.create(ctx, req)
		return err
	})
	return resp, err
}

func (ps *paymentService) updatePaymentStatus(
	ctx context.Context,
	payment entity.Payment,
	newStatus entity.PaymentStatus,
	statusErr error,
	startsAt, endsAt time.Time,
) error {
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
		payment.StartsAt = sql.NullTime{
			Time:  startsAt,
			Valid: true,
		}
		payment.EndsAt = sql.NullTime{
			Time:  endsAt,
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
