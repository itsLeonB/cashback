package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/subscription"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type SubscriptionService interface {
	// Admin
	Create(ctx context.Context, req dto.NewSubscriptionRequest) (dto.SubscriptionResponse, error)
	GetList(ctx context.Context) ([]dto.SubscriptionResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)
	Update(ctx context.Context, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)

	// Public
	CreatePurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error)

	// Internal
	AttachDefaultSubscription(ctx context.Context, profileID uuid.UUID) error
	GetCurrentSubscription(ctx context.Context, profileID uuid.UUID) (entity.Subscription, error)
	TransitionStatus(ctx context.Context, msg message.SubscriptionStatusTransitioned) error
}

type subscriptionService struct {
	transactor       crud.Transactor
	subscriptionRepo crud.Repository[entity.Subscription]
	planVersionRepo  crud.Repository[entity.PlanVersion]
	paymentService   PaymentService
}

func NewSubscriptionService(
	transactor crud.Transactor,
	repo crud.Repository[entity.Subscription],
	planVersionRepo crud.Repository[entity.PlanVersion],
	paymentService PaymentService,
) *subscriptionService {
	return &subscriptionService{
		transactor,
		repo,
		planVersionRepo,
		paymentService,
	}
}

func (ss *subscriptionService) Create(ctx context.Context, req dto.NewSubscriptionRequest) (dto.SubscriptionResponse, error) {
	newSubscription := entity.Subscription{
		ProfileID:     req.ProfileID,
		PlanVersionID: req.PlanVersionID,
		AutoRenew:     req.AutoRenew,
	}

	if !req.EndsAt.IsZero() {
		newSubscription.EndsAt = sql.NullTime{
			Time:  req.EndsAt,
			Valid: true,
		}
	}
	if !req.CanceledAt.IsZero() {
		newSubscription.CanceledAt = sql.NullTime{
			Time:  req.CanceledAt,
			Valid: true,
		}
	}

	insertedSubscription, err := ss.subscriptionRepo.Insert(ctx, newSubscription)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return mapper.SubscriptionToResponse(insertedSubscription), nil
}

func (ss *subscriptionService) GetList(ctx context.Context) ([]dto.SubscriptionResponse, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.PreloadRelations = []string{"Profile", "PlanVersion.Plan"}
	subscriptions, err := ss.subscriptionRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(subscriptions, mapper.SubscriptionToResponse), nil
}

func (ss *subscriptionService) GetOne(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	subscription, err := ss.getByID(ctx, id, false)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return mapper.SubscriptionToResponse(subscription), nil
}

func (ss *subscriptionService) Update(ctx context.Context, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	var resp dto.SubscriptionResponse
	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ss.getByID(ctx, req.ID, true)
		if err != nil {
			return err
		}

		subscription.ProfileID = req.ProfileID
		subscription.PlanVersionID = req.PlanVersionID
		subscription.AutoRenew = req.AutoRenew

		subscription.EndsAt = sql.NullTime{
			Time:  req.EndsAt,
			Valid: !req.EndsAt.IsZero(),
		}

		subscription.CanceledAt = sql.NullTime{
			Time:  req.CanceledAt,
			Valid: !req.CanceledAt.IsZero(),
		}

		updatedSubscription, err := ss.subscriptionRepo.Update(ctx, subscription)
		if err != nil {
			return err
		}

		resp = mapper.SubscriptionToResponse(updatedSubscription)
		return nil
	})
	return resp, err
}

func (ss *subscriptionService) Delete(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	var resp dto.SubscriptionResponse
	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ss.getByID(ctx, id, true)
		if err != nil {
			return err
		}

		if err = ss.subscriptionRepo.Delete(ctx, subscription); err != nil {
			return err
		}

		resp = mapper.SubscriptionToResponse(subscription)
		return nil
	})
	return resp, err
}

func (ss *subscriptionService) AttachDefaultSubscription(ctx context.Context, profileID uuid.UUID) error {
	planVerSpec := crud.Specification[entity.PlanVersion]{}
	planVerSpec.Model.IsDefault = true
	planVersion, err := ss.planVersionRepo.FindFirst(ctx, planVerSpec)
	if err != nil {
		return err
	}
	if planVersion.IsZero() {
		return ungerr.Unknown("no default plan version is found")
	}

	newSubsReq := dto.NewSubscriptionRequest{
		ProfileID:     profileID,
		PlanVersionID: planVersion.ID,
	}
	_, err = ss.Create(ctx, newSubsReq)
	return err
}

func (ss *subscriptionService) GetCurrentSubscription(ctx context.Context, profileID uuid.UUID) (entity.Subscription, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.Model.ProfileID = profileID
	spec.PreloadRelations = []string{"PlanVersion", "PlanVersion.Plan"}

	subscriptions, err := ss.subscriptionRepo.FindAll(ctx, spec)
	if err != nil {
		return entity.Subscription{}, err
	}

	now := time.Now()
	for _, sub := range subscriptions {
		if sub.IsActive(now) {
			return sub, nil
		}
	}

	return entity.Subscription{}, nil
}

func (ss *subscriptionService) CreatePurchase(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.PaymentResponse, error) {
	if err := ss.paymentService.IsReady(); err != nil {
		return dto.PaymentResponse{}, err
	}

	var resp dto.PaymentResponse
	var err error
	err = ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		planVerSpec := crud.Specification[entity.PlanVersion]{}
		planVerSpec.Model.ID = req.PlanVersionID
		planVerSpec.Model.PlanID = req.PlanID
		planVersion, err := ss.planVersionRepo.FindFirst(ctx, planVerSpec)
		if err != nil {
			return err
		}
		if planVersion.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("plan version ID %s is not found", req.PlanVersionID))
		}

		subsSpec := crud.Specification[entity.Subscription]{}
		subsSpec.Model.ProfileID = req.ProfileID
		subsSpec.Model.PlanVersionID = req.PlanVersionID
		existingSubs, err := ss.subscriptionRepo.FindAll(ctx, subsSpec)
		if err != nil {
			return err
		}
		for _, sub := range existingSubs {
			if sub.Status == entity.SubscriptionActive || sub.Status == entity.SubscriptionPastDuePayment {
				return ungerr.ConflictError("user still have existing subscription")
			}
		}

		newSubscription := entity.Subscription{
			ProfileID:     req.ProfileID,
			PlanVersionID: req.PlanVersionID,
			Status:        entity.SubscriptionIncompletePayment,
		}

		insertedSubs, err := ss.subscriptionRepo.Insert(ctx, newSubscription)
		if err != nil {
			return err
		}

		newPayment := dto.NewPaymentRequest{
			SubscriptionID: insertedSubs.ID,
			Amount:         planVersion.PriceAmount,
			Currency:       planVersion.PriceCurrency,
		}

		resp, err = ss.paymentService.Create(ctx, newPayment)
		return err
	})
	return resp, err
}

func (ss *subscriptionService) TransitionStatus(ctx context.Context, msg message.SubscriptionStatusTransitioned) error {
	return ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[entity.Subscription]{}
		spec.Model.ID = msg.ID
		spec.ForUpdate = true
		spec.PreloadRelations = []string{"Payments"}
		subs, err := ss.subscriptionRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if subs.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("subscription ID %s is not found", msg.ID))
		}

		patchedSubs, err := subscription.TransitionStatus(subs, msg.Status)
		if err != nil {
			return err
		}

		_, err = ss.subscriptionRepo.Update(ctx, patchedSubs)
		return err
	})
}

func (ss *subscriptionService) getByID(ctx context.Context, id uuid.UUID, forUpdate bool) (entity.Subscription, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	spec.PreloadRelations = []string{"Profile", "PlanVersion.Plan"}
	subscription, err := ss.subscriptionRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Subscription{}, err
	}
	if subscription.IsZero() {
		return entity.Subscription{}, ungerr.NotFoundError("subscription is not found")
	}
	return subscription, nil
}
