package handler

import (
	"time"

	"github.com/itsLeonB/cashback/internal/adapters/http/cookie"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
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
	Public                *PublicHandler
}

func (h *Handlers) Shutdown() {
	h.Auth.emailLimiter.Stop()
}

func ProvideHandlers(services *provider.Services, cookieCfg cookie.Config) *Handlers {
	return &Handlers{
		NewAuthHandler(services.Auth, services.OAuth, services.Session, cookieCfg, middlewares.NewValueLimiter(3.0/3600, 3, time.Hour)),
		NewFriendshipHandler(services.Friendship, services.FriendDetails, services.Debt),
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
		NewPublicHandler(services.FriendDetails),
	}
}
