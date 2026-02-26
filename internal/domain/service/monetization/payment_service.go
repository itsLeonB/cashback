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
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PaymentService interface {
	NewPurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error)
	HandleNotification(ctx context.Context, req dto.MidtransNotificationPayload) error
	MakePayment(ctx context.Context, subscriptionID uuid.UUID) (dto.PaymentResponse, error)

	// Admin
	GetList(ctx context.Context) ([]dto.PaymentResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error)
	Update(ctx context.Context, req dto.UpdatePaymentRequest) (dto.PaymentResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error)
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

func (ps *paymentService) isReady() error {
	if !config.Global.SubscriptionPurchaseEnabled {
		return ungerr.ForbiddenError("feature is disabled")
	}
	if ps.gateway == nil {
		return ungerr.Unknown("payment gateway is uninitialized")
	}
	return nil
}

func (ps *paymentService) NewPurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error) {
	if err := ps.isReady(); err != nil {
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
		ExpiredAt: sql.NullTime{
			Time:  time.Now().Add(24 * time.Hour),
			Valid: true,
		},
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
	if err := ps.isReady(); err != nil {
		return err
	}

	newStatus, statusErr := ps.gateway.CheckStatus(ctx, req)
	if statusErr != nil {
		if newStatus == "" {
			return statusErr
		}
		logger.Error(statusErr)
	}

	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		id, err := ezutil.Parse[uuid.UUID](req.OrderID)
		if err != nil {
			return err
		}

		// 1. Fetch payment to extract SubscriptionID (no lock)
		specInfo := crud.Specification[entity.Payment]{}
		specInfo.Model.ID = id
		specInfo.Model.Gateway = ps.gateway.Provider()
		paymentInfo, err := ps.paymentRepo.FindFirst(ctx, specInfo)
		if err != nil {
			return err
		}
		if paymentInfo.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("payment with ID %s is not found", id))
		}

		// 2. Lock Subscription FIRST to prevent deadlocks with MakePayment
		subs, err := ps.subscriptionSvc.GetByID(ctx, paymentInfo.SubscriptionID, true)
		if err != nil {
			return err
		}

		// 3. Lock Payment
		specUpdate := specInfo
		specUpdate.ForUpdate = true
		payment, err := ps.paymentRepo.FindFirst(ctx, specUpdate)
		if err != nil {
			return err
		}
		if !payment.IsSettleable() || newStatus == payment.Status {
			return nil
		}

		startsAt, endsAt := subs.ContinuedPeriods()

		if err = ps.updatePaymentStatus(ctx, payment, newStatus, statusErr, startsAt, endsAt); err != nil {
			return err
		}

		return ps.updateSubscriptionStatus(ctx, subs, newStatus, startsAt, endsAt)
	})
}

func (ps *paymentService) updateSubscriptionStatus(
	ctx context.Context,
	subs entity.Subscription,
	newStatus entity.PaymentStatus,
	startsAt, endsAt time.Time,
) error {
	switch newStatus {
	case entity.PaidPayment:
		subs.Status = entity.SubscriptionActive
		subs.CurrentPeriodStart = sql.NullTime{Time: startsAt, Valid: true}
		subs.CurrentPeriodEnd = sql.NullTime{Time: endsAt, Valid: true}
		return ps.subscriptionSvc.Save(ctx, subs)
	case entity.ErrorPayment, entity.CanceledPayment, entity.ExpiredPayment:
		if subs.Status == entity.SubscriptionIncompletePayment {
			subs.Status = entity.SubscriptionCanceled
			return ps.subscriptionSvc.Save(ctx, subs)
		}
	}
	return nil
}

func (ps *paymentService) MakePayment(ctx context.Context, subscriptionID uuid.UUID) (dto.PaymentResponse, error) {
	if err := ps.isReady(); err != nil {
		return dto.PaymentResponse{}, err
	}

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ps.subscriptionSvc.GetByID(ctx, subscriptionID, true)
		if err != nil {
			return err
		}

		if subscription.Status == entity.SubscriptionCanceled {
			return ungerr.ForbiddenError("cannot make payment for canceled subscription")
		}

		// Check for existing pending/processing payments
		spec := crud.Specification[entity.Payment]{}
		spec.Model.SubscriptionID = subscriptionID
		payments, err := ps.paymentRepo.FindAll(ctx, spec)
		if err != nil {
			return err
		}

		for _, p := range payments {
			if p.Status == entity.PendingPayment || p.Status == entity.ProcessingPayment {
				// We found an incomplete payment
				if p.ExpiredAt.Valid && p.ExpiredAt.Time.Before(time.Now()) {
					// It's expired, mark it as such to free up the unique constraint
					p.Status = entity.ExpiredPayment
					if _, err := ps.paymentRepo.Update(ctx, p); err != nil {
						return err
					}
				} else {
					// It's still valid, return idempotently
					resp = mapper.PaymentToResponse(p)
					return nil
				}
			}
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

func (ps *paymentService) GetList(ctx context.Context) ([]dto.PaymentResponse, error) {
	payments, err := ps.paymentRepo.FindAll(ctx, crud.Specification[entity.Payment]{})
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(payments, mapper.PaymentToResponse), nil
}

func (ps *paymentService) GetOne(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error) {
	payment, err := ps.getByID(ctx, id)
	if err != nil {
		return dto.PaymentResponse{}, err
	}

	return mapper.PaymentToResponse(payment), nil
}

func (ps *paymentService) Update(ctx context.Context, req dto.UpdatePaymentRequest) (dto.PaymentResponse, error) {
	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := ps.getByID(ctx, req.ID)
		if err != nil {
			return err
		}

		payment.Status = entity.PaymentStatus(req.Status)
		payment.Amount = req.Amount
		payment.Currency = req.Currency

		payment.StartsAt = sql.NullTime{
			Time:  req.StartsAt,
			Valid: !req.StartsAt.IsZero(),
		}
		payment.EndsAt = sql.NullTime{
			Time:  req.EndsAt,
			Valid: !req.EndsAt.IsZero(),
		}
		payment.PaidAt = sql.NullTime{
			Time:  req.PaidAt,
			Valid: !req.PaidAt.IsZero(),
		}

		updatedPayment, err := ps.paymentRepo.Update(ctx, payment)
		if err != nil {
			return err
		}

		resp = mapper.PaymentToResponse(updatedPayment)
		return nil
	})
	return resp, err
}

func (ps *paymentService) Delete(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error) {
	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := ps.getByID(ctx, id)
		if err != nil {
			return err
		}

		if err = ps.paymentRepo.Delete(ctx, payment); err != nil {
			return err
		}

		resp = mapper.PaymentToResponse(payment)
		return nil
	})
	return resp, err
}

func (ps *paymentService) getByID(ctx context.Context, id uuid.UUID) (entity.Payment, error) {
	spec := crud.Specification[entity.Payment]{}
	spec.Model.ID = id
	payment, err := ps.paymentRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Payment{}, err
	}
	if payment.IsZero() {
		return entity.Payment{}, ungerr.NotFoundError("payment is not found")
	}
	return payment, nil
}
