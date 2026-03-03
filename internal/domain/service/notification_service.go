package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
)

type notificationService struct {
	repo         repository.NotificationRepository
	debtSvc      DebtService
	friendReqSvc FriendshipRequestService
	friendSvc    FriendshipService
	expenseSvc   GroupExpenseService
	taskQueue    queue.TaskQueue
}

func NewNotificationService(
	repo repository.NotificationRepository,
	debtSvc DebtService,
	friendReqSvc FriendshipRequestService,
	friendSvc FriendshipService,
	expenseSvc GroupExpenseService,
	taskQueue queue.TaskQueue,
) *notificationService {
	return &notificationService{
		repo,
		debtSvc,
		friendReqSvc,
		friendSvc,
		expenseSvc,
		taskQueue,
	}
}

func (ns *notificationService) HandleDebtCreated(ctx context.Context, msg message.DebtCreated) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.HandleDebtCreated")
	defer span.End()

	return ns.publishNotification(ctx, func(ctx context.Context) (entity.Notification, error) {
		return ns.debtSvc.ConstructNotification(ctx, msg)
	})
}

func (ns *notificationService) HandleFriendRequestSent(ctx context.Context, msg message.FriendRequestSent) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.HandleFriendRequestSent")
	defer span.End()

	return ns.publishNotification(ctx, func(ctx context.Context) (entity.Notification, error) {
		return ns.friendReqSvc.ConstructNotification(ctx, msg)
	})
}

func (ns *notificationService) HandleFriendRequestAccepted(ctx context.Context, msg message.FriendRequestAccepted) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.HandleFriendRequestAccepted")
	defer span.End()

	return ns.publishNotification(ctx, func(ctx context.Context) (entity.Notification, error) {
		return ns.friendSvc.ConstructNotification(ctx, msg)
	})
}

func (ns *notificationService) HandleExpenseConfirmed(ctx context.Context, msg message.ExpenseConfirmed) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.HandleExpenseConfirmed")
	defer span.End()

	notifications, err := ns.expenseSvc.ConstructNotifications(ctx, msg)
	if err != nil {
		return err
	}

	createdNotifs, err := ns.repo.CreateMany(ctx, notifications)
	if err != nil {
		return err
	}

	go func() {
		for _, createdNotif := range createdNotifs {
			ns.taskQueue.AsyncEnqueue(ctx, message.NotificationCreated{ID: createdNotif.ID})
		}
	}()

	return nil
}

func (ns *notificationService) GetUnread(ctx context.Context, profileID uuid.UUID) ([]dto.NotificationResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.GetUnread")
	defer span.End()

	notifications, err := ns.repo.GetByProfileID(ctx, profileID, true)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(notifications, mapper.NotificationToResponse), nil
}

func (ns *notificationService) MarkAsRead(ctx context.Context, profileID, notificationID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.MarkAsRead")
	defer span.End()

	return ns.repo.MarkAsRead(ctx, profileID, notificationID)
}

func (ns *notificationService) MarkAllAsRead(ctx context.Context, profileID uuid.UUID) error {
	ctx, span := otel.Tracer.Start(ctx, "NotificationService.MarkAllAsRead")
	defer span.End()

	return ns.repo.MarkAllAsRead(ctx, profileID)
}

func (ns *notificationService) publishNotification(ctx context.Context, constructorFn func(ctx context.Context) (entity.Notification, error)) error {
	notification, err := constructorFn(ctx)
	if err != nil {
		return err
	}

	createdNotif, err := ns.repo.New(ctx, notification)
	if err != nil {
		return err
	}

	go ns.taskQueue.AsyncEnqueue(ctx, message.NotificationCreated{ID: createdNotif.ID})
	return nil
}
