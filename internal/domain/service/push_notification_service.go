package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/webpush"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/mapper/notification"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/datatypes"
)

type pushNotificationService struct {
	repo             crud.Repository[entity.PushSubscription]
	notificationRepo repository.NotificationRepository
	webPushClient    webpush.Client
}

func NewPushNotificationService(
	repo crud.Repository[entity.PushSubscription],
	notificationRepo repository.NotificationRepository,
	webPushClient webpush.Client,
) *pushNotificationService {
	return &pushNotificationService{
		repo,
		notificationRepo,
		webPushClient,
	}
}

func (s *pushNotificationService) Subscribe(ctx context.Context, req dto.PushSubscriptionRequest) error {
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

func (s *pushNotificationService) Unsubscribe(ctx context.Context, req dto.PushUnsubscribeRequest) error {
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

func (s *pushNotificationService) Deliver(ctx context.Context, msg message.NotificationCreated) error {
	spec := crud.Specification[entity.Notification]{}
	spec.Model.ID = msg.ID
	notif, err := s.notificationRepo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if notif.ID == uuid.Nil {
		return nil
	}

	// Skip if notification is read (soft-deleted equivalent)
	if notif.ReadAt.Valid {
		return nil
	}

	title, err := notification.ResolveTitle(notif)
	if err != nil {
		logger.Error(err)
		return nil
	}

	// Get all push subscriptions for the profile
	subscriptionSpec := crud.Specification[entity.PushSubscription]{}
	subscriptionSpec.Model.ProfileID = notif.ProfileID
	subscriptions, err := s.repo.FindAll(ctx, subscriptionSpec)
	if err != nil {
		return err
	}

	// No subscriptions - silent no-op
	if len(subscriptions) == 0 {
		return nil
	}

	// Construct push payload
	payload := map[string]interface{}{
		"title": title,
		"data": map[string]interface{}{
			"notification_id": msg.ID.String(),
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return ungerr.Wrap(err, "failed to marshal push payload")
	}

	// Send to all subscriptions
	for _, subscription := range subscriptions {
		keys, err := ezutil.Unmarshal[webpush.Keys](subscription.Keys)
		if err != nil {
			logger.Errorf("error unmarshaling key for subscription %s: %v", subscription.ID, err)
			continue
		}

		if err := s.webPushClient.Send(webpush.Subscription{
			Endpoint: subscription.Endpoint,
			Keys:     keys,
			Payload:  payloadBytes,
		}); err != nil {
			logger.Errorf("failed to send push to subscription %s: %v", subscription.ID, err)
		}
	}

	return nil
}
