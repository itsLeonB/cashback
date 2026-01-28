package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
)

type notificationService struct {
	repo    repository.NotificationRepository
	debtSvc DebtService
}

func NewNotificationService(
	repo repository.NotificationRepository,
	debtSvc DebtService,
) *notificationService {
	return &notificationService{
		repo,
		debtSvc,
	}
}

func (ns *notificationService) HandleDebtCreated(ctx context.Context, msg message.DebtCreated) error {
	notification, err := ns.debtSvc.ConstructNotification(ctx, msg)
	if err != nil {
		return err
	}

	_, err = ns.repo.New(ctx, notification)
	return err
}

func (ns *notificationService) GetUnread(ctx context.Context, profileID uuid.UUID) ([]dto.NotificationResponse, error) {
	notifications, err := ns.repo.GetByProfileID(ctx, profileID, true)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(notifications, mapper.NotificationToResponse), nil
}

func (ns *notificationService) MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error {
	return ns.repo.MarkAsRead(ctx, profileID, notificationID)
}

func (ns *notificationService) MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error {
	return ns.repo.MarkAllAsRead(ctx, profileID)
}
