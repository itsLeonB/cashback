package provider

import (
	"net/http"

	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/fee"
	"github.com/itsLeonB/sekure"
)

type Services struct {
	// Auth
	Auth service.AuthService

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

	// Infra
	Notification service.NotificationService
}

func (s *Services) Shutdown() error {
	return s.TransferMethod.Shutdown()
}

func ProvideServices(
	repos *Repositories,
	coreSvc *CoreServices,
	authConfig config.Auth,
	appConfig config.App,
) *Services {
	profile := service.NewProfileService(repos.Transactor, repos.Profile, repos.User, repos.Friendship, repos.RelatedProfile)
	friendship := service.NewFriendshipService(repos.Transactor, repos.Friendship, profile)
	friendReq := service.NewFriendshipRequestService(repos.Transactor, friendship, profile, repos.FriendshipRequest, coreSvc.Queue)

	groupExpense := service.NewGroupExpenseService(friendship, repos.GroupExpense, repos.Transactor, fee.NewFeeCalculatorRegistry(), repos.OtherFee, repos.ExpenseBill, coreSvc.LLM, coreSvc.Image, coreSvc.Queue)

	transferMethod := service.NewTransferMethodService(repos.TransferMethod, coreSvc.Storage, appConfig.BucketNameTransferMethods, appembed.TransferMethodAssets)
	debt := service.NewDebtService(repos.DebtTransaction, transferMethod, friendship, profile, groupExpense, coreSvc.Queue)

	return &Services{
		Auth: provideAuth(authConfig, repos, profile, appConfig, coreSvc),

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

		Notification: service.NewNotificationService(repos.Notification, debt, friendReq, friendship, groupExpense),
	}
}

func provideAuth(
	authConfig config.Auth,
	repos *Repositories,
	profile service.ProfileService,
	appConfig config.App,
	coreSvc *CoreServices,
) service.AuthService {
	jwt := sekure.NewJwtService(authConfig.Issuer, authConfig.SecretKey, authConfig.TokenDuration)
	user := service.NewUserService(repos.Transactor, repos.User, profile, repos.PasswordResetToken)

	return service.NewAuthService(
		jwt,
		repos.Transactor,
		user,
		coreSvc.Mail,
		appConfig.RegisterVerificationUrl,
		appConfig.ResetPasswordUrl,
		service.NewOAuthService(repos.Transactor, repos.OAuthAccount, coreSvc.State, user, http.DefaultClient, jwt),
		authConfig.HashCost,
	)
}
