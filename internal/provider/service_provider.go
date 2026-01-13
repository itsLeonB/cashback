package provider

import (
	"net/http"

	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/debt"
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
}

func (s *Services) Shutdown() error {
	return s.TransferMethod.Shutdown()
}

func ProvideServices(
	repos *Repositories,
	coreSvc *CoreServices,
	queues *Queues,
	authConfig config.Auth,
	appConfig config.App,
) *Services {
	jwt := sekure.NewJwtService(authConfig.Issuer, authConfig.SecretKey, authConfig.TokenDuration)

	profile := service.NewProfileService(repos.Transactor, repos.Profile, repos.User, repos.Friendship, repos.RelatedProfile)
	user := service.NewUserService(repos.Transactor, repos.User, profile, repos.PasswordResetToken)
	friendship := service.NewFriendshipService(repos.Transactor, repos.Friendship, profile)

	transferMethod := service.NewTransferMethodService(repos.TransferMethod, coreSvc.Storage, appConfig.BucketNameTransferMethods, appembed.TransferMethodAssets)
	debt := service.NewDebtService(debt.NewDebtCalculatorStrategies(), repos.DebtTransaction, transferMethod, friendship, profile)

	expenseBill := service.NewExpenseBillService(appConfig.BucketNameExpenseBill, queues.taskQueue, repos.ExpenseBill, repos.Transactor, coreSvc.Image, coreSvc.OCR)
	groupExpense := service.NewGroupExpenseService(friendship, repos.GroupExpense, repos.Transactor, fee.NewFeeCalculatorRegistry(), repos.OtherFee, repos.ExpenseBill, coreSvc.LLM, expenseBill, debt)

	return &Services{
		Auth: service.NewAuthService(
			jwt,
			repos.Transactor,
			user,
			coreSvc.Mail,
			appConfig.RegisterVerificationUrl,
			appConfig.ResetPasswordUrl,
			service.NewOAuthService(repos.Transactor, repos.OAuthAccount, coreSvc.State, user, http.DefaultClient, jwt),
			authConfig.HashCost,
		),

		Profile:           profile,
		Friendship:        friendship,
		FriendshipRequest: service.NewFriendshipRequestService(repos.Transactor, friendship, profile, repos.FriendshipRequest),
		FriendDetails:     service.NewFriendDetailsService(debt, profile, friendship),

		Debt:                  debt,
		TransferMethod:        transferMethod,
		ProfileTransferMethod: service.NewProfileTransferMethodService(profile, repos.ProfileTransferMethod, transferMethod, friendship),

		GroupExpense: groupExpense,
		ExpenseBill:  expenseBill,
		ExpenseItem:  service.NewExpenseItemService(repos.Transactor, repos.GroupExpense, repos.ExpenseItem, groupExpense),
		OtherFee:     service.NewOtherFeeService(repos.Transactor, repos.GroupExpense, repos.OtherFee, groupExpense),
	}
}
