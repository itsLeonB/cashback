package monetization

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	entity "github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	mapper "github.com/itsLeonB/cashback/internal/domain/mapper/monetization"
	"github.com/itsLeonB/cashback/internal/domain/message"
	repository "github.com/itsLeonB/cashback/internal/domain/repository/monetization"
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
	GetSubscribedDetails(ctx context.Context, profileID uuid.UUID) (dto.SubscriptionResponse, error)

	// Internal
	AttachDefaultSubscription(ctx context.Context, profileID uuid.UUID) error
	GetCurrentSubscription(ctx context.Context, profileID uuid.UUID, isActive bool) (entity.Subscription, error)
	TransitionStatus(ctx context.Context, msg message.SubscriptionStatusTransitioned) error
	CreateNew(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.NewPaymentRequest, error)
	GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (entity.Subscription, error)
	UpdatePastDues(ctx context.Context) error
	PublishSubscriptionDueNotifications(ctx context.Context) error
}

type subscriptionService struct {
	transactor       crud.Transactor
	subscriptionRepo repository.SubscriptionRepository
	planVersionRepo  crud.Repository[entity.PlanVersion]
	taskQueue        queue.TaskQueue
}

func NewSubscriptionService(
	transactor crud.Transactor,
	repo repository.SubscriptionRepository,
	planVersionRepo crud.Repository[entity.PlanVersion],
	taskQueue queue.TaskQueue,
) *subscriptionService {
	return &subscriptionService{
		transactor,
		repo,
		planVersionRepo,
		taskQueue,
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

	return mapper.SubscriptionToResponse(insertedSubscription, time.Now()), nil
}

func (ss *subscriptionService) GetList(ctx context.Context) ([]dto.SubscriptionResponse, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.PreloadRelations = []string{"Profile", "PlanVersion.Plan"}
	subscriptions, err := ss.subscriptionRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(subscriptions, mapper.SimpleSubscriptionMapper()), nil
}

func (ss *subscriptionService) GetOne(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	subscription, err := ss.GetByID(ctx, id, false)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return mapper.SubscriptionToResponse(subscription, time.Now()), nil
}

func (ss *subscriptionService) Update(ctx context.Context, req dto.UpdateSubscriptionRequest) (dto.SubscriptionResponse, error) {
	var resp dto.SubscriptionResponse
	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ss.GetByID(ctx, req.ID, true)
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

		resp = mapper.SubscriptionToResponse(updatedSubscription, time.Now())
		return nil
	})
	return resp, err
}

func (ss *subscriptionService) Delete(ctx context.Context, id uuid.UUID) (dto.SubscriptionResponse, error) {
	var resp dto.SubscriptionResponse
	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		subscription, err := ss.GetByID(ctx, id, true)
		if err != nil {
			return err
		}

		if err = ss.subscriptionRepo.Delete(ctx, subscription); err != nil {
			return err
		}

		resp = mapper.SubscriptionToResponse(subscription, time.Now())
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

func (ss *subscriptionService) GetCurrentSubscription(ctx context.Context, profileID uuid.UUID, isActive bool) (entity.Subscription, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.Model.ProfileID = profileID
	spec.PreloadRelations = []string{"PlanVersion", "PlanVersion.Plan"}

	subscriptions, err := ss.subscriptionRepo.FindAll(ctx, spec)
	if err != nil {
		return entity.Subscription{}, err
	}

	sort.Slice(subscriptions, func(i, j int) bool {
		return subscriptions[i].PlanVersion.Plan.Priority < subscriptions[j].PlanVersion.Plan.Priority
	})

	now := time.Now()
	for _, sub := range subscriptions {
		if isActive {
			if sub.IsActive(now) {
				return sub, nil
			}
		} else {
			if sub.IsSubscribed(now) {
				return sub, nil
			}
		}
	}

	return entity.Subscription{}, nil
}

func (ss *subscriptionService) CreateNew(ctx context.Context, req dto.PurchaseSubscriptionRequest) (dto.NewPaymentRequest, error) {
	var resp dto.NewPaymentRequest
	err := ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
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

		var newSubscription entity.Subscription
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
			if sub.Status == entity.SubscriptionIncompletePayment {
				newSubscription = sub
			}
		}

		var subID uuid.UUID
		if newSubscription.IsZero() {
			newSubscription = entity.Subscription{
				ProfileID:     req.ProfileID,
				PlanVersionID: req.PlanVersionID,
				Status:        entity.SubscriptionIncompletePayment,
			}
			insertedSubs, err := ss.subscriptionRepo.Insert(ctx, newSubscription)
			if err != nil {
				return err
			}
			subID = insertedSubs.ID
		} else {
			subID = newSubscription.ID
		}

		resp = dto.NewPaymentRequest{
			SubscriptionID: subID,
			Amount:         planVersion.PriceAmount,
			Currency:       planVersion.PriceCurrency,
		}

		return nil
	})
	return resp, err
}

func (ss *subscriptionService) TransitionStatus(ctx context.Context, msg message.SubscriptionStatusTransitioned) error {
	return ss.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[entity.Subscription]{}
		spec.Model.ID = msg.ID
		spec.ForUpdate = true
		spec.PreloadRelations = []string{"Payments", "PlanVersion"}
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

func (ss *subscriptionService) GetSubscribedDetails(ctx context.Context, profileID uuid.UUID) (dto.SubscriptionResponse, error) {
	sub, err := ss.GetCurrentSubscription(ctx, profileID, false)
	if err != nil {
		return dto.SubscriptionResponse{}, err
	}

	return mapper.SubscriptionToResponse(sub, time.Now()), nil
}

func (ss *subscriptionService) GetByID(ctx context.Context, id uuid.UUID, forUpdate bool) (entity.Subscription, error) {
	spec := crud.Specification[entity.Subscription]{}
	spec.Model.ID = id
	spec.ForUpdate = forUpdate
	spec.PreloadRelations = []string{"Profile", "PlanVersion", "PlanVersion.Plan"}
	subscription, err := ss.subscriptionRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Subscription{}, err
	}
	if subscription.IsZero() {
		return entity.Subscription{}, ungerr.NotFoundError("subscription is not found")
	}
	return subscription, nil
}

func (ss *subscriptionService) UpdatePastDues(ctx context.Context) error {
	return ss.subscriptionRepo.UpdatePastDues(ctx)
}

func (ss *subscriptionService) PublishSubscriptionDueNotifications(ctx context.Context) error {
	subscriptions, err := ss.subscriptionRepo.FindNearingDueDate(ctx)
	if err != nil {
		return err
	}

	userIDs := ezutil.MapSlice(subscriptions, func(sub entity.Subscription) uuid.UUID {
		return sub.Profile.UserID.UUID
	})

	if len(userIDs) == 0 {
		logger.Infof("no subscriptions nearing payment due date")
		return nil
	}

	logger.Infof("%d subscriptions nearing payment due date", len(userIDs))

	msg := message.SubscriptionNearingDue{
		UserIDs: userIDs,
	}

	return ss.taskQueue.Enqueue(ctx, msg)
}
