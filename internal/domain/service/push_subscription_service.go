package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/datatypes"
)

type pushSubscriptionService struct {
	repo crud.Repository[entity.PushSubscription]
}

func NewPushSubscriptionService(repo crud.Repository[entity.PushSubscription]) *pushSubscriptionService {
	return &pushSubscriptionService{repo}
}

func (s *pushSubscriptionService) Subscribe(ctx context.Context, req dto.PushSubscriptionRequest) error {
	// Check if subscription already exists
	spec := crud.Specification[entity.PushSubscription]{}
	spec.Model.ProfileID = req.ProfileID
	spec.Model.Endpoint = req.Endpoint
	existing, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	// If exists, update is handled by unique constraint on endpoint
	if existing.ID != uuid.Nil {
		return nil // Gracefully handle re-subscription
	}

	keys := entity.PushSubscriptionKeys{
		P256dh: req.Keys.P256dh,
		Auth:   req.Keys.Auth,
	}

	keysJSON, err := json.Marshal(keys)
	if err != nil {
		return ungerr.Wrap(err, "failed to marshal keys")
	}

	subscription := entity.PushSubscription{
		ID:        uuid.New(),
		ProfileID: req.ProfileID,
		Endpoint:  req.Endpoint,
		Keys:      datatypes.JSON(keysJSON),
	}

	if req.UserAgent != "" {
		subscription.UserAgent = sql.NullString{
			String: req.UserAgent,
			Valid:  true,
		}
	}

	_, err = s.repo.Insert(ctx, subscription)
	return err
}

func (s *pushSubscriptionService) Unsubscribe(ctx context.Context, req dto.PushUnsubscribeRequest) error {
	spec := crud.Specification[entity.PushSubscription]{}
	spec.Model.ProfileID = req.ProfileID
	spec.Model.Endpoint = req.Endpoint
	existing, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if existing.ID == uuid.Nil {
		return nil
	}

	return s.repo.Delete(ctx, existing)
}
