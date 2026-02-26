package handler

import (
	"github.com/itsLeonB/cashback/internal/provider"
)

type Handlers struct {
	Auth                  *AuthHandler
	Friendship            *FriendshipHandler
	FriendshipRequest     *FriendshipRequestHandler
	Profile               *ProfileHandler
	TransferMethod        *TransferMethodHandler
	Debt                  *DebtHandler
	GroupExpense          *groupExpenseHandler
	ExpenseItem           *ExpenseItemHandler
	OtherFee              *OtherFeeHandler
	ExpenseBill           *ExpenseBillHandler
	ProfileTransferMethod *ProfileTransferMethodHandler
	Notification          *NotificationHandler
	PushSubscription      *PushSubscriptionHandler
	Subscription          *SubscriptionHandler
	Payment               *PaymentHandler
	Plan                  *PlanHandler
}

func ProvideHandlers(services *provider.Services) *Handlers {
	return &Handlers{
		NewAuthHandler(services.Auth, services.OAuth, services.Session),
		NewFriendshipHandler(services.Friendship, services.FriendDetails),
		NewFriendshipRequestHandler(services.FriendshipRequest),
		NewProfileHandler(services.Profile),
		NewTransferMethodHandler(services.TransferMethod),
		NewDebtHandler(services.Debt),
		newGroupExpenseHandler(services.GroupExpense),
		NewExpenseItemHandler(services.ExpenseItem),
		NewOtherFeeHandler(services.OtherFee),
		NewExpenseBillHandler(services.ExpenseBill),
		&ProfileTransferMethodHandler{services.ProfileTransferMethod},
		NewNotificationHandler(services.Notification),
		NewPushSubscriptionHandler(services.PushNotification),
		&SubscriptionHandler{services.Subscription, services.Payment},
		&PaymentHandler{services.Payment},
		&PlanHandler{services.PlanVersion},
	}
}
