package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/mapper/notification"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type pushDeliveryService struct {
	pushSubscriptionRepo crud.Repository[entity.PushSubscription]
	notificationRepo     repository.NotificationRepository
	transactor           crud.Transactor
	pushConfig           config.Push
}

func NewPushDeliveryService(
	pushSubscriptionRepo crud.Repository[entity.PushSubscription],
	notificationRepo repository.NotificationRepository,
	transactor crud.Transactor,
	pushConfig config.Push,
) PushDeliveryService {
	return &pushDeliveryService{
		pushSubscriptionRepo,
		notificationRepo,
		transactor,
		pushConfig,
	}
}

func (s *pushDeliveryService) DeliverToProfile(ctx context.Context, msg message.NotificationCreated) error {
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

	_, err = notification.ResolveTitle(notif)
	if err != nil {
		logger.Error(err)
		return nil
	}

	// TODO: Implement actual push delivery using VAPID keys and push subscriptions
	// For now, this is a placeholder that would:
	// 1. Get all push subscriptions for the profile
	// 2. Send push notification to each subscription endpoint
	// 3. Handle delivery failures appropriately

	return ungerr.Unknown("push delivery not implemented")
}
