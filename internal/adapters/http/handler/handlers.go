package handler

import (
	"context"
	"time"

	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/go-authkit/authgin"
)

type Handlers struct {
	Auth                  *authgin.Handler
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

	emailLimiter *middlewares.ValueLimiter
}

func (h *Handlers) Shutdown() {
	h.emailLimiter.Stop()
}

func ProvideHandlers(services *provider.Services, authCfg config.Auth) *Handlers {
	transport := authgin.NewCookieTransport(authgin.CookieConfig{
		Domain:     authCfg.CookieDomain,
		Secure:     authCfg.CookieSecure,
		SameSite:   authCfg.ParsedSameSite(),
		AccessTTL:  authCfg.TokenDuration,
		RefreshTTL: authCfg.RefreshTokenDuration,
	})

	emailLimiter := middlewares.NewValueLimiter(3.0/3600, 3, time.Hour)

	authHandler := authgin.NewHandler(services.AuthKit, transport, authgin.HandlerConfig{
		Captcha: &captchaAdapter{inner: services.Captcha},
		Limiter: emailLimiter,
	})

	return &Handlers{
		Auth:                  authHandler,
		Friendship:            NewFriendshipHandler(services.Friendship, services.FriendDetails, services.Debt),
		FriendshipRequest:     NewFriendshipRequestHandler(services.FriendshipRequest),
		Profile:               NewProfileHandler(services.Profile),
		TransferMethod:        NewTransferMethodHandler(services.TransferMethod),
		Debt:                  NewDebtHandler(services.Debt),
		GroupExpense:          newGroupExpenseHandler(services.GroupExpense),
		ExpenseItem:           NewExpenseItemHandler(services.ExpenseItem),
		OtherFee:              NewOtherFeeHandler(services.OtherFee),
		ExpenseBill:           NewExpenseBillHandler(services.ExpenseBill),
		ProfileTransferMethod: &ProfileTransferMethodHandler{services.ProfileTransferMethod},
		Notification:          NewNotificationHandler(services.Notification),
		PushSubscription:      NewPushSubscriptionHandler(services.PushNotification),
		Subscription:          &SubscriptionHandler{services.Subscription, services.Payment},
		Payment:               &PaymentHandler{services.Payment},
		Plan:                  &PlanHandler{services.PlanVersion},
		Public:                NewPublicHandler(services.FriendDetails),
		emailLimiter:          emailLimiter,
	}
}

// captchaAdapter adapts service.CaptchaService to authgin.CaptchaVerifier.
type captchaAdapter struct {
	inner service.CaptchaService
}

func (a *captchaAdapter) Verify(ctx context.Context, token string) error {
	return a.inner.Verify(ctx, token)
}
