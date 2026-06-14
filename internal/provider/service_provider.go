package provider

import (
	"errors"

	"github.com/google/uuid"
	appembed "github.com/itsLeonB/cashback"
	authadapter "github.com/itsLeonB/cashback/internal/adapters/auth"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/cache"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/fee"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	"github.com/itsLeonB/go-authkit"
	"github.com/markbates/goth/providers/google"
)

type Services struct {
	// Auth
	AuthKit *authkit.AuthKit
	Captcha service.CaptchaService

	// Users
	User              service.UserService
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
	Payment      monetization.PaymentService

	// Infra
	Notification     service.NotificationService
	PushNotification service.PushNotificationService
}

func (s *Services) Shutdown() error {
	return errors.Join(s.AuthKit.Shutdown(), s.TransferMethod.Shutdown())
}

func ProvideServices(
	repos *Repositories,
	coreSvc *CoreServices,
) *Services {
	authConfig := config.Global.Auth
	appConfig := config.Global.App
	oauthConfig := config.Global.OAuthProviders
	paymentConfig := config.Global.Payment

	paymentGateway, err := payment.NewGateway(paymentConfig)
	if err != nil {
		logger.Error(err)
	}

	subs := monetization.NewSubscriptionService(repos.Transactor, repos.Subscription, repos.PlanVersion, coreSvc.Queue)
	payment := monetization.NewPaymentService(paymentGateway, repos.Transactor, repos.Payment, coreSvc.Queue, subs)
	subsLimit := service.NewSubscriptionLimitService(subs, repos.ExpenseBill)

	profile := service.NewProfileService(repos.Transactor, repos.Profile, repos.User, repos.Friendship, repos.RelatedProfile, subs, subsLimit)
	user := service.NewUserService(repos.Transactor, repos.User, profile, repos.PasswordResetToken, coreSvc.Mail)
	friendship := service.NewFriendshipService(repos.Transactor, repos.Friendship, profile)
	pushNotification := service.NewPushNotificationService(repos.PushSubscription, repos.Notification, repos.Transactor, coreSvc.WebPush)

	// Auth adapters bridge authkit store interfaces to existing repos/infra.
	txAdapter := authadapter.NewTransactor(repos.Transactor)
	userStore := authadapter.NewUserStore(repos.User, profile)
	sessionStore := authadapter.NewSessionStore(repos.Session)
	refreshTokenStore := authadapter.NewRefreshTokenStore(repos.RefreshToken)
	resetTokenStore := authadapter.NewResetTokenStore(repos.PasswordResetToken)
	oauthAccountStore := authadapter.NewOAuthAccountStore(repos.OAuthAccount)
	mailAdapter := authadapter.NewMailAdapter(coreSvc.Mail)
	sessionCache := cache.NewInMemoryCache[uuid.UUID](authConfig.TokenDuration)
	cacheAdapter := authadapter.NewSessionCacheAdapter(sessionCache)
	stateAdapter := authadapter.NewStateStore(coreSvc.State)

	hooks := NewAuthKitHooks(pushNotification, profile, friendship)

	kit := authkit.New(authkit.Config{
		VerificationURL:  appConfig.RegisterVerificationUrl,
		ResetPasswordURL: appConfig.ResetPasswordUrl,
		RefreshTokenTTL:  authConfig.RefreshTokenDuration,
		JWTIssuer:        authConfig.Issuer,
		JWTSecret:        authConfig.SecretKey,
		JWTDuration:      authConfig.TokenDuration,
	}, authkit.Deps{
		Tx:       txAdapter,
		Users:    userStore,
		Sessions: sessionStore,
		Refresh:  refreshTokenStore,
		Resets:   resetTokenStore,
		OAuth:    oauthAccountStore,
		Mail:     mailAdapter,
		Cache:    cacheAdapter,
		State:    stateAdapter,
		Providers: []authkit.ProviderConfig{
			{Provider: google.New(oauthConfig.Google.ClientID, oauthConfig.Google.ClientSecret, oauthConfig.Google.RedirectUrl, "email", "profile"), Trusted: true},
		},
	}, hooks)

	friendReq := service.NewFriendshipRequestService(repos.Transactor, friendship, profile, repos.FriendshipRequest, coreSvc.Queue)

	groupExpense := service.NewGroupExpenseService(friendship, repos.GroupExpense, repos.Transactor, fee.NewFeeCalculatorRegistry(), repos.OtherFee, repos.ExpenseBill, coreSvc.LLM, coreSvc.Image, coreSvc.Queue, coreSvc.Langfuse, profile)

	transferMethod := service.NewTransferMethodService(repos.TransferMethod, coreSvc.Storage, appConfig.BucketNameTransferMethods, appembed.TransferMethodAssets)
	debt := service.NewDebtService(repos.DebtTransaction, transferMethod, friendship, profile, groupExpense, coreSvc.Queue)

	return &Services{
		AuthKit: kit,
		Captcha: service.NewTurnstileService(authConfig.TurnstileSecretKey),

		User:              user,
		Profile:           profile,
		Friendship:        friendship,
		FriendshipRequest: friendReq,
		FriendDetails:     service.NewFriendDetailsService(debt, profile, friendship),

		Debt:                  debt,
		TransferMethod:        transferMethod,
		ProfileTransferMethod: service.NewProfileTransferMethodService(profile, repos.ProfileTransferMethod, transferMethod, friendship),

		GroupExpense: groupExpense,
		ExpenseBill:  service.NewExpenseBillService(coreSvc.Queue, repos.ExpenseBill, repos.Transactor, coreSvc.Image, coreSvc.OCR, groupExpense, subsLimit),
		ExpenseItem:  service.NewExpenseItemService(repos.Transactor, repos.ExpenseItem, groupExpense),
		OtherFee:     service.NewOtherFeeService(repos.Transactor, repos.GroupExpense, repos.OtherFee, groupExpense),

		Plan:         monetization.NewPlanService(repos.Transactor, repos.Plan, repos.PlanVersion),
		PlanVersion:  monetization.NewPlanVersionService(repos.Transactor, repos.PlanVersion),
		Subscription: subs,
		Payment:      payment,

		Notification:     service.NewNotificationService(repos.Notification, debt, friendReq, friendship, groupExpense, coreSvc.Queue),
		PushNotification: pushNotification,
	}
}
