package monetization

import (
	"context"

	"github.com/google/uuid"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type SubscriptionService interface {
	Create(ctx context.Context, req dto.NewSubscriptionRequest) (dto.SubscriptionResponse, error)
	GetList(ctx context.Context) ([]dto.SubscriptionResponse, error)
	GetOne(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)
	Update(ctx context.Context, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error)
	Delete(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error)
}

type subscriptionService struct {
	transactor       crud.Transactor
	subscriptionRepo crud.Repository[entity.Subscription]
}

func NewSubscriptionService(
	transactor crud.Transactor,
	repo crud.Repository[entity.Subscription],
) *subscriptionService {
	return &subscriptionService{
		transactor,
		repo,
	}
}

func (ss *subscriptionService) Create(ctx context.Context, req dto.NewSubscriptionRequest) (dto.SubscriptionResponse, error) {
	newSubscription := entity.Subscription{
		ProfileID:     req.ProfileID,
		PlanVersionID: req.PlanVersionID,
		EndsAt:        req.EndsAt,
		CanceledAt:    req.CanceledAt,
		AutoRenew:     req.AutoRenew,
	}

	insertedSubscription, err := ss.subscriptionRepo.Insert(ctx, newSubscription)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	subscription, err := ss.getByID(ctx, insertedSubscription.ID, false)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return mapper.SubscriptionToResponse(subscription), nil
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
		subscription.EndsAt = req.EndsAt
		subscription.CanceledAt = req.CanceledAt
		subscription.AutoRenew = req.AutoRenew

		updatedSubscription, err := ss.subscriptionRepo.Update(ctx, subscription)
		if err != nil {
			return err
		}

		reloaded, err := ss.getByID(ctx, updatedSubscription.ID, false)
		if err != nil {
			return err
		}

		resp = mapper.SubscriptionToResponse(reloaded)
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
