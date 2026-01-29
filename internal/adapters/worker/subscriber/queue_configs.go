package subscriber

import (
	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/adapters/worker/subscriber/handler"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/provider"
)

type queueConfig struct {
	name     string
	handler  asynq.Handler
	priority int
}

func configureQueues(providers *provider.Providers) ([]queueConfig, map[string]int) {
	queues := []queueConfig{
		{
			message.ExpenseBillUploaded{}.Type(),
			withLogging(message.ExpenseBillUploaded{}.Type(), providers.Services.ExpenseBill.ExtractBillText),
			3,
		},
		{
			message.ExpenseBillTextExtracted{}.Type(),
			withLogging(message.ExpenseBillTextExtracted{}.Type(), providers.Services.GroupExpense.ParseFromBillText),
			3,
		},
		{
			message.ExpenseConfirmed{}.Type(),
			withLogging(message.ExpenseConfirmed{}.Type(),
				handler.ExpenseConfirmedHandler(
					providers.Debt,
					providers.Services.Notification,
					providers.Services.GroupExpense,
				)),
			3,
		},
		{
			message.DebtCreated{}.Type(),
			withLogging(message.DebtCreated{}.Type(), providers.Services.Notification.HandleDebtCreated),
			3,
		},
		{
			message.FriendRequestSent{}.Type(),
			withLogging(message.FriendRequestSent{}.Type(), providers.Services.Notification.HandleFriendRequestSent),
			3,
		},
		{
			message.FriendRequestAccepted{}.Type(),
			withLogging(message.FriendRequestAccepted{}.Type(), providers.Services.Notification.HandleFriendRequestAccepted),
			3,
		},
		{
			message.NotificationCreated{}.Type(),
			withLogging(message.NotificationCreated{}.Type(), providers.PushDelivery.DeliverToProfile),
			3,
		},
	}

	queuePriorities := make(map[string]int, len(queues))
	for _, q := range queues {
		queuePriorities[q.name] = q.priority
	}

	return queues, queuePriorities
}
