package subscriber

import (
	"github.com/itsLeonB/cashback/internal/adapters/worker/subscriber/handler"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/nats-io/nats.go/jetstream"
)

type queueConfig struct {
	name    string
	handler jetstream.MessageHandler
}

func configureQueues(providers *provider.Providers) []queueConfig {
	return []queueConfig{
		{
			message.ExpenseBillUploaded{}.Type(),
			withLogging(message.ExpenseBillUploaded{}.Type(), providers.Services.ExpenseBill.ExtractBillText),
		},
		{
			message.ExpenseBillTextExtracted{}.Type(),
			withLogging(message.ExpenseBillTextExtracted{}.Type(), providers.Services.GroupExpense.ParseFromBillText),
		},
		{
			message.ExpenseConfirmed{}.Type(),
			withLogging[message.ExpenseConfirmed](message.ExpenseConfirmed{}.Type(),
				handler.ExpenseConfirmedHandler(
					providers.Debt,
					providers.Services.Notification,
					providers.Services.GroupExpense,
				)),
		},
		{
			message.DebtCreated{}.Type(),
			withLogging(message.DebtCreated{}.Type(), providers.Services.Notification.HandleDebtCreated),
		},
		{
			message.FriendRequestSent{}.Type(),
			withLogging(message.FriendRequestSent{}.Type(), providers.Services.Notification.HandleFriendRequestSent),
		},
		{
			message.FriendRequestAccepted{}.Type(),
			withLogging(message.FriendRequestAccepted{}.Type(), providers.Services.Notification.HandleFriendRequestAccepted),
		},
		{
			message.NotificationCreated{}.Type(),
			withLogging(message.NotificationCreated{}.Type(), providers.PushNotification.Deliver),
		},
		{
			message.SubscriptionNearingDue{}.Type(),
			withLogging(message.SubscriptionNearingDue{}.Type(), providers.Services.User.SendSubscriptionNearingDueDateMail),
		},
	}
}

// compile-time check
var _ queue.TaskMessage = message.ExpenseBillUploaded{}
