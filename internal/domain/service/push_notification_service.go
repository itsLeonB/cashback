package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"sync"
	"time"

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
	repo             repository.PushSubscriptionRepository
	notificationRepo repository.NotificationRepository
	transactor       crud.Transactor
	webPushClient    webpush.Client
}

func NewPushNotificationService(
	repo repository.PushSubscriptionRepository,
	notificationRepo repository.NotificationRepository,
	transactor crud.Transactor,
	webPushClient webpush.Client,
) *pushNotificationService {
	return &pushNotificationService{
		repo,
		notificationRepo,
		transactor,
		webPushClient,
	}
}

func (s *pushNotificationService) Subscribe(ctx context.Context, req dto.PushSubscriptionRequest) error {
	keysJSON, err := json.Marshal(entity.PushSubscriptionKeys{
		P256dh: req.Keys.P256dh,
		Auth:   req.Keys.Auth,
	})
	if err != nil {
		return ungerr.Wrap(err, "failed to marshal keys")
	}

	return s.repo.Upsert(ctx, entity.PushSubscription{
		ProfileID: req.ProfileID,
		Endpoint:  req.Endpoint,
		Keys:      datatypes.JSON(keysJSON),
		UserAgent: sql.NullString{
			String: req.UserAgent,
			Valid:  req.UserAgent != "",
		},
	})
}

func (s *pushNotificationService) Unsubscribe(ctx context.Context, req dto.PushUnsubscribeRequest) error {
	spec := crud.Specification[entity.PushSubscription]{}
	spec.Model.ProfileID = req.ProfileID
	spec.Model.Endpoint = req.Endpoint
	existing, err := s.repo.FindFirst(ctx, spec)
	if err != nil {
		return err
	}
	if existing.IsZero() {
		return nil
	}

	return s.repo.Delete(ctx, existing)
}

func (s *pushNotificationService) Deliver(ctx context.Context, msg message.NotificationCreated) error {
	return s.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		notif, err := s.getPushableNotification(ctx, msg.ID)
		if err != nil {
			return err
		}
		if notif.IsZero() {
			return nil
		}

		if err = s.deliverToSubs(ctx, notif); err != nil {
			return err
		}

		notif.PushedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}

		_, err = s.notificationRepo.Update(ctx, notif)
		if err != nil {
			logger.Error(err)
		}

		return nil
	})
}

func (s *pushNotificationService) deliverToSubs(ctx context.Context, notif entity.Notification) error {
	title, err := notification.ResolveTitle(notif)
	if err != nil {
		logger.Error(err)
		logger.Warn("using default notification title")
		title = "Notification"
	}

	// Get all push subscriptions for the profile
	spec := crud.Specification[entity.PushSubscription]{}
	spec.Model.ProfileID = notif.ProfileID
	subscriptions, err := s.repo.FindAll(ctx, spec)
	if err != nil {
		return err
	}

	// No subscriptions - silent no-op
	if len(subscriptions) == 0 {
		logger.Warnf("profileID %s has no subscriptions", notif.ProfileID)
		return nil
	}

	// Construct push payload
	payloadBytes, err := json.Marshal(
		map[string]interface{}{
			"title": title,
			"data": map[string]interface{}{
				"notification_id": notif.ID.String(),
			},
		},
	)
	if err != nil {
		return ungerr.Wrap(err, "failed to marshal push payload")
	}

	var wg sync.WaitGroup
	// Send to all subscriptions
	for _, subscription := range subscriptions {
		wg.Go(func() {
			s.sendSubscription(subscription, payloadBytes)
		})
	}
	wg.Wait()
	return nil
}

func (s *pushNotificationService) sendSubscription(subscription entity.PushSubscription, payload []byte) {
	keys, err := ezutil.Unmarshal[webpush.Keys](subscription.Keys)
	if err != nil {
		logger.Errorf("error unmarshaling key for subscription %s: %v", subscription.ID, err)
		return
	}

	if err := s.webPushClient.Send(webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys:     keys,
		Payload:  payload,
	}); err != nil {
		logger.Errorf("failed to send push to subscription %s: %v", subscription.ID, err)
	}
}

func (s *pushNotificationService) getPushableNotification(ctx context.Context, id uuid.UUID) (entity.Notification, error) {
	spec := crud.Specification[entity.Notification]{}
	spec.Model.ID = id
	spec.ForUpdate = true
	notif, err := s.notificationRepo.FindFirst(ctx, spec)
	if err != nil {
		return entity.Notification{}, err
	}
	if notif.IsZero() {
		logger.Errorf("notification ID: %s is not found", id)
		return entity.Notification{}, nil
	}
	// Skip if notification is read/pushed
	if notif.ReadAt.Valid || notif.PushedAt.Valid {
		return entity.Notification{}, nil
	}
	return notif, nil
}
