package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/mapper/notification"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/datatypes"
)

type pushNotificationService struct {
	repo             crud.Repository[entity.PushSubscription]
	notificationRepo repository.NotificationRepository
	pushConfig       config.Push
}

func NewPushNotificationService(
	repo crud.Repository[entity.PushSubscription],
	notificationRepo repository.NotificationRepository,
	pushConfig config.Push,
) *pushNotificationService {
	return &pushNotificationService{
		repo,
		notificationRepo,
		pushConfig,
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
	// Skip if push notifications are disabled
	if !s.pushConfig.Enabled {
		return nil
	}

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
		if err := s.sendPushNotification(subscription, payloadBytes); err != nil {
			// Log individual failures but don't fail the job
			logger.Errorf("failed to send push to subscription %s: %v", subscription.ID, err)
		}
	}

	return nil
}

func (s *pushNotificationService) sendPushNotification(subscription entity.PushSubscription, payload []byte) error {
	// Unmarshal keys from JSONB
	var keys entity.PushSubscriptionKeys
	if err := json.Unmarshal(subscription.Keys, &keys); err != nil {
		return ungerr.Wrap(err, "failed to unmarshal subscription keys")
	}

	// Create webpush subscription
	webpushSub := &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: keys.P256dh,
			Auth:   keys.Auth,
		},
	}

	// Send push notification
	resp, err := webpush.SendNotification(payload, webpushSub, &webpush.Options{
		VAPIDPublicKey:  s.pushConfig.VapidPublicKey,
		VAPIDPrivateKey: s.pushConfig.VapidPrivateKey,
		Subscriber:      s.pushConfig.VapidSubject,
	})
	if err != nil {
		return err
	}
	defer func() {
		if e := resp.Body.Close(); e != nil {
			logger.Errorf("error closing response body: %v", e)
		}
	}()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ungerr.Unknownf("push service returned status %d", resp.StatusCode)
	}

	return nil
}
