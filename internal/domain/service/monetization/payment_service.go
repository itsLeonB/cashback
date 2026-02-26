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
	if err := ps.IsReady(); err != nil {
		return err
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

		// 4. Calculate Billing Period safely from CurrentPeriodEnd
		startsAt := time.Now()
		if subs.Status == entity.SubscriptionActive && subs.CurrentPeriodEnd.Valid && subs.CurrentPeriodEnd.Time.After(startsAt) {
			startsAt = subs.CurrentPeriodEnd.Time
		}
		endsAt := startsAt
		switch subs.PlanVersion.BillingInterval {
		case entity.MonthlyInterval:
			endsAt = endsAt.AddDate(0, 1, 0)
		case entity.YearlyInterval:
			endsAt = endsAt.AddDate(1, 0, 0)
		}

		// 5. Save Payment
		if err = ps.updatePaymentStatus(ctx, payment, newStatus, err, startsAt, endsAt); err != nil {
			return err
		}

		// 6. Inline State Mutation
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
