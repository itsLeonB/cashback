package provider

import (
	"net/http"

	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/fee"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/sekure"
)

type Services struct {
	// Auth
	Auth    service.AuthService
	OAuth   service.OAuthService
	Session service.SessionService

	// Users
	Profile           service.ProfileService
	Friendship        service.FriendshipService
	FriendshipRequest service.FriendshipRequestService
	FriendDetails     service.FriendDetailsService

	// Debts
	Debt                  service.DebtService
	TransferMethod        service.TransferMethodService
	ProfileTransferMethod service.ProfileTransferMethodService

	// Expenses
	GroupExpense service.GroupExpenseService
	ExpenseBill  service.ExpenseBillService
	ExpenseItem  service.ExpenseItemService
	OtherFee     service.OtherFeeService

	// Monetization
	Plan         monetization.PlanService
	PlanVersion  monetization.PlanVersionService
	Subscription monetization.SubscriptionService

	// Infra
	Notification     service.NotificationService
	PushNotification service.PushNotificationService
}

func (s *Services) Shutdown() error {
	return s.TransferMethod.Shutdown()
}

func ProvideServices(
	repos *Repositories,
	coreSvc *CoreServices,
	authConfig config.Auth,
	appConfig config.App,
	pushConfig config.Push,
) *Services {
	jwt := sekure.NewJwtService(authConfig.Issuer, authConfig.SecretKey, authConfig.TokenDuration)
	profile := service.NewProfileService(repos.Transactor, repos.Profile, repos.User, repos.Friendship, repos.RelatedProfile, repos.Subscription)
	user := service.NewUserService(repos.Transactor, repos.User, profile, repos.PasswordResetToken)
	session := service.NewSessionService(jwt, user, repos.Transactor, repos.Session, repos.RefreshToken)

	friendship := service.NewFriendshipService(repos.Transactor, repos.Friendship, profile)
	friendReq := service.NewFriendshipRequestService(repos.Transactor, friendship, profile, repos.FriendshipRequest, coreSvc.Queue)

	groupExpense := service.NewGroupExpenseService(friendship, repos.GroupExpense, repos.Transactor, fee.NewFeeCalculatorRegistry(), repos.OtherFee, repos.ExpenseBill, coreSvc.LLM, coreSvc.Image, coreSvc.Queue)

	transferMethod := service.NewTransferMethodService(repos.TransferMethod, coreSvc.Storage, appConfig.BucketNameTransferMethods, appembed.TransferMethodAssets)
	debt := service.NewDebtService(repos.DebtTransaction, transferMethod, friendship, profile, groupExpense, coreSvc.Queue)

	pushNotification := service.NewPushNotificationService(repos.PushSubscription, repos.Notification, repos.Transactor, coreSvc.WebPush)

	return &Services{
		Auth:    service.NewAuthService(jwt, repos.Transactor, user, coreSvc.Mail, appConfig.RegisterVerificationUrl, appConfig.ResetPasswordUrl, authConfig.HashCost, pushNotification, session),
		OAuth:   service.NewOAuthService(repos.Transactor, repos.OAuthAccount, coreSvc.State, user, http.DefaultClient, session),
		Session: session,

		Profile:           profile,
		Friendship:        friendship,
		FriendshipRequest: friendReq,
		FriendDetails:     service.NewFriendDetailsService(debt, profile, friendship),

		Debt:                  debt,
		TransferMethod:        transferMethod,
		ProfileTransferMethod: service.NewProfileTransferMethodService(profile, repos.ProfileTransferMethod, transferMethod, friendship),

		GroupExpense: groupExpense,
		ExpenseBill:  service.NewExpenseBillService(coreSvc.Queue, repos.ExpenseBill, repos.Transactor, coreSvc.Image, coreSvc.OCR, groupExpense),
		ExpenseItem:  service.NewExpenseItemService(repos.Transactor, repos.ExpenseItem, groupExpense),
		OtherFee:     service.NewOtherFeeService(repos.Transactor, repos.GroupExpense, repos.OtherFee, groupExpense),

		Plan:         monetization.NewPlanService(repos.Transactor, repos.Plan, repos.PlanVersion),
		PlanVersion:  monetization.NewPlanVersionService(repos.Transactor, repos.PlanVersion, repos.Plan),
		Subscription: monetization.NewSubscriptionService(repos.Transactor, repos.Subscription),

		Notification:     service.NewNotificationService(repos.Notification, debt, friendReq, friendship, groupExpense, coreSvc.Queue),
		PushNotification: pushNotification,
	}
}
