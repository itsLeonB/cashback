package service

import (
	"context"

	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
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
