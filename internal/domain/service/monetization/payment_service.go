package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/otel"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type PaymentService interface {
	NewPurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error)
	HandleWebhook(ctx context.Context, payload []byte, signature string) error
	CreatePortalSession(ctx context.Context, profileID uuid.UUID) (dto.PortalSessionResponse, error)

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
	profileRepo crud.Repository[users.UserProfile],
	userGetter UserGetter,
	subscriptionSvc SubscriptionService,
) *paymentService {
	return &paymentService{
		gateway:         gateway,
		transactor:      transactor,
		paymentRepo:     paymentRepo,
		profileRepo:     profileRepo,
		userGetter:      userGetter,
		subscriptionSvc: subscriptionSvc,
	}
}

// UserGetter retrieves a user by ID (avoids circular import with service package).
type UserGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (users.User, error)
}

type paymentService struct {
	gateway         payment.Gateway
	transactor      crud.Transactor
	paymentRepo     crud.Repository[entity.Payment]
	profileRepo     crud.Repository[users.UserProfile]
	userGetter      UserGetter
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
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.NewPurchase")
	defer span.End()

	if err := ps.isReady(); err != nil {
		return dto.PaymentResponse{}, err
	}

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		paymentRequest, planVersion, err := ps.subscriptionSvc.CreateNew(ctx, req)
		if err != nil {
			return err
		}

		// Get profile for email and stripe customer ID
		profileSpec := crud.Specification[users.UserProfile]{}
		profileSpec.Model.ID = req.ProfileID
		profile, err := ps.profileRepo.FindFirst(ctx, profileSpec)
		if err != nil {
			return err
		}
		if profile.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("profile ID %s is not found", req.ProfileID))
		}

		// Look up user email via profile's user
		if !profile.UserID.Valid {
			return ungerr.ForbiddenError("anonymous profiles cannot make purchases")
		}
		user, err := ps.userGetter.GetByID(ctx, profile.UserID.UUID)
		if err != nil {
			return err
		}

		newPayment := entity.Payment{
			SubscriptionID: paymentRequest.SubscriptionID,
			Amount:         paymentRequest.Amount,
			Currency:       paymentRequest.Currency,
			Status:         entity.PendingPayment,
			Gateway:        "stripe",
		}

		pendingPayment, err := ps.paymentRepo.Insert(ctx, newPayment)
		if err != nil {
			return err
		}

		result, err := ps.gateway.CreateCheckoutSession(payment.CheckoutParams{
			Payment:      pendingPayment,
			PlanVersion:  planVersion,
			CustomerID:   profile.StripeCustomerID.String,
			ProfileEmail: user.Email,
			ProfileID:    req.ProfileID,
			SuccessURL:   config.Global.SuccessURL,
			CancelURL:    config.Global.CancelURL,
		})
		if err != nil {
			return err
		}

		pendingPayment.GatewayTransactionID = sql.NullString{String: result.GatewaySessionID, Valid: true}
		pendingPayment, err = ps.paymentRepo.Update(ctx, pendingPayment)
		if err != nil {
			return err
		}

		// Persist stripe customer ID if new
		if !profile.StripeCustomerID.Valid && result.GatewayCustomerID != "" {
			profile.StripeCustomerID = sql.NullString{String: result.GatewayCustomerID, Valid: true}
			if _, err := ps.profileRepo.Update(ctx, profile); err != nil {
				return err
			}
		}

		paymentResp := mapper.PaymentToResponse(pendingPayment)
		paymentResp.CheckoutURL = result.CheckoutURL
		resp = paymentResp
		return nil
	})
	return resp, err
}

func (ps *paymentService) HandleWebhook(ctx context.Context, payload []byte, signature string) error {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.HandleWebhook")
	defer span.End()

	if ps.gateway == nil {
		return ungerr.Unknown("payment gateway is uninitialized")
	}

	event, err := ps.gateway.HandleWebhook(payload, signature)
	if err != nil {
		return err
	}
	if event == nil {
		return nil
	}

	// Idempotency check (only for payment-related events)
	if event.Type == "payment_success" || event.Type == "payment_failed" || event.Type == "subscription_canceled" {
		spec := crud.Specification[entity.Payment]{}
		spec.Model.GatewayEventID = sql.NullString{String: event.GatewayEventID, Valid: true}
		existing, err := ps.paymentRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if !existing.IsZero() {
			return nil
		}
	}

	return ps.updateSubscriptionByEvent(ctx, event)
}

func (ps *paymentService) updateSubscriptionByEvent(ctx context.Context, event *payment.WebhookEvent) error {
	return ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		sub, err := ps.subscriptionSvc.GetByID(ctx, event.SubscriptionID, true)
		if err != nil {
			return err
		}

		switch event.Type {
		case "payment_success":
			sub, err = ps.handlePaymentSuccess(ctx, event, sub)
		case "payment_failed":
			sub.Status = entity.SubscriptionPastDuePayment

		case "subscription_canceled":
			sub.Status = entity.SubscriptionCanceled
			if !sub.CanceledAt.Valid {
				sub.CanceledAt = sql.NullTime{Time: time.Now(), Valid: true}
			}

		case "subscription_scheduled_cancel":
			sub.CanceledAt = sql.NullTime{Time: event.CancelAt, Valid: true}
			sub.EndsAt = sql.NullTime{Time: event.CancelAt, Valid: true}

		case "subscription_reactivated":
			sub.CanceledAt = sql.NullTime{}
			sub.EndsAt = sql.NullTime{}

		}
		if err != nil {
			return err
		}

		return ps.subscriptionSvc.Save(ctx, sub)
	})
}

func (ps *paymentService) handlePaymentSuccess(ctx context.Context, event *payment.WebhookEvent, sub entity.Subscription) (entity.Subscription, error) {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.handlePaymentSuccess")
	defer span.End()

	// Find pending payment for this subscription
	spec := crud.Specification[entity.Payment]{}
	spec.Model.SubscriptionID = event.SubscriptionID
	spec.Model.Status = entity.PendingPayment
	spec.ForUpdate = true
	p, err := ps.paymentRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Subscription{}, err
	}
	if p.IsZero() {
		// No pending payment, might be a renewal — create a record
		p = entity.Payment{
			SubscriptionID: event.SubscriptionID,
			Amount:         sub.PlanVersion.PriceAmount,
			Currency:       sub.PlanVersion.PriceCurrency,
			Gateway:        "stripe",
		}
		p, err = ps.paymentRepo.Insert(ctx, p)
		if err != nil {
			return entity.Subscription{}, err
		}
	}

	now := time.Now()
	p.Status = entity.PaidPayment
	p.PaidAt = sql.NullTime{Time: now, Valid: true}
	p.GatewayEventID = sql.NullString{String: event.GatewayEventID, Valid: true}
	p.GatewaySubscriptionID = sql.NullString{String: event.GatewaySubID, Valid: true}

	periodStart := event.PeriodStart
	periodEnd := event.PeriodEnd
	if periodStart.IsZero() {
		periodStart = now
	}
	if periodEnd.IsZero() {
		_, periodEnd = sub.ContinuedPeriods()
	}

	p.StartsAt = sql.NullTime{Time: periodStart, Valid: true}
	p.EndsAt = sql.NullTime{Time: periodEnd, Valid: true}

	if _, err := ps.paymentRepo.Update(ctx, p); err != nil {
		return entity.Subscription{}, err
	}

	sub.Status = entity.SubscriptionActive
	sub.CurrentPeriodStart = sql.NullTime{Time: periodStart, Valid: true}
	sub.CurrentPeriodEnd = sql.NullTime{Time: periodEnd, Valid: true}

	if event.GatewaySubID != "" {
		sub.GatewaySubscriptionID = sql.NullString{String: event.GatewaySubID, Valid: true}
	}

	return sub, nil
}

func (ps *paymentService) CreatePortalSession(ctx context.Context, profileID uuid.UUID) (dto.PortalSessionResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.CreatePortalSession")
	defer span.End()

	if ps.gateway == nil {
		return dto.PortalSessionResponse{}, ungerr.Unknown("payment gateway is uninitialized")
	}

	profileSpec := crud.Specification[users.UserProfile]{}
	profileSpec.Model.ID = profileID
	profile, err := ps.profileRepo.FindFirst(ctx, profileSpec)
	if err != nil {
		return dto.PortalSessionResponse{}, err
	}
	if profile.IsZero() {
		return dto.PortalSessionResponse{}, ungerr.NotFoundError(fmt.Sprintf("profile ID %s not found", profileID))
	}
	if !profile.StripeCustomerID.Valid {
		return dto.PortalSessionResponse{}, ungerr.NotFoundError("no stripe customer found for this profile")
	}

	url, err := ps.gateway.CreatePortalSession(profile.StripeCustomerID.String, config.Global.SuccessURL)
	if err != nil {
		return dto.PortalSessionResponse{}, err
	}

	return dto.PortalSessionResponse{PortalURL: url}, nil
}

func (ps *paymentService) GetList(ctx context.Context) ([]dto.PaymentResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.GetList")
	defer span.End()

	payments, err := ps.paymentRepo.FindAll(ctx, crud.Specification[entity.Payment]{})
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(payments, mapper.PaymentToResponse), nil
}

func (ps *paymentService) GetOne(ctx context.Context, id uuid.UUID) (dto.PaymentResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.GetOne")
	defer span.End()

	payment, err := ps.getByID(ctx, id, false)
	if err != nil {
		return dto.PaymentResponse{}, err
	}

	return mapper.PaymentToResponse(payment), nil
}

func (ps *paymentService) Update(ctx context.Context, req dto.UpdatePaymentRequest) (dto.PaymentResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.Update")
	defer span.End()

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := ps.getByID(ctx, req.ID, true)
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
	ctx, span := otel.Tracer.Start(ctx, "PaymentService.Delete")
	defer span.End()

	var resp dto.PaymentResponse
	err := ps.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		payment, err := ps.getByID(ctx, id, true)
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

func (ps *paymentService) getByID(ctx context.Context, id uuid.UUID, forUpdate bool) (entity.Payment, error) {
	spec := crud.Specification[entity.Payment]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	payment, err := ps.paymentRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Payment{}, err
	}
	if payment.IsZero() {
		return entity.Payment{}, ungerr.NotFoundError("payment is not found")
	}
	return payment, nil
}
