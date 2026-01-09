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
	Auth  service.AuthService
	OAuth service.OAuthService

	// Users
	User              service.UserService
	Profile           service.ProfileService
	Friendship        service.FriendshipService
	FriendshipRequest service.FriendshipRequestService
	FriendDetails     service.FriendDetailsService

	// Debts
	Debt           service.DebtService
	TransferMethod service.TransferMethodService

	// Expenses
	GroupExpense service.GroupExpenseService
	ExpenseBill  service.ExpenseBillService
	ExpenseItem  service.ExpenseItemService
	OtherFee     service.OtherFeeService
}

func ProvideServices(
	repos *Repositories,
	coreSvc *CoreServices,
	queues *Queues,
	authConfig config.Auth,
	appConfig config.App,
) *Services {
	hash := sekure.NewHashService(authConfig.HashCost)
	jwt := sekure.NewJwtService(authConfig.Issuer, authConfig.SecretKey, authConfig.TokenDuration)

	profile := service.NewProfileService(repos.Transactor, repos.Profile, repos.User, repos.Friendship, repos.RelatedProfile)
	user := service.NewUserService(repos.Transactor, repos.User, profile, repos.PasswordResetToken)

	oauth := service.NewOAuthService(repos.Transactor, repos.OAuthAccount, coreSvc.State, user, http.DefaultClient, jwt)
	auth := service.NewAuthService(hash, jwt, repos.Transactor, user, coreSvc.Mail, appConfig.RegisterVerificationUrl, appConfig.ResetPasswordUrl, oauth)

	friendship := service.NewFriendshipService(repos.Transactor, repos.Friendship, profile)
	friendshipReq := service.NewFriendshipRequestService(repos.Transactor, friendship, profile, repos.FriendshipRequest)

	transferMethod := service.NewTransferMethodService(repos.TransferMethod, coreSvc.Storage, appConfig.BucketNameTransferMethods, appembed.TransferMethodAssets)
	debt := service.NewDebtService(debt.NewDebtCalculatorStrategies(), repos.DebtTransaction, transferMethod, friendship, profile)
	friendDetail := service.NewFriendDetailsService(debt, profile, friendship)

	expenseBill := service.NewExpenseBillService(appConfig.BucketNameExpenseBill, queues.taskQueue, repos.ExpenseBill, repos.Transactor, coreSvc.Image, coreSvc.OCR)
	groupExpense := service.NewGroupExpenseService(friendship, repos.GroupExpense, repos.Transactor, fee.NewFeeCalculatorRegistry(), repos.OtherFee, repos.ExpenseBill, coreSvc.LLM, expenseBill, debt)
	expenseItem := service.NewExpenseItemService(repos.Transactor, repos.GroupExpense, repos.ExpenseItem, groupExpense)
	otherFee := service.NewOtherFeeService(repos.Transactor, repos.GroupExpense, repos.OtherFee, groupExpense)

	return &Services{
		Auth:              auth,
		OAuth:             oauth,
		User:              user,
		Profile:           profile,
		Friendship:        friendship,
		FriendshipRequest: friendshipReq,
		FriendDetails:     friendDetail,
		Debt:              debt,
		TransferMethod:    transferMethod,
		GroupExpense:      groupExpense,
		ExpenseBill:       expenseBill,
		ExpenseItem:       expenseItem,
		OtherFee:          otherFee,
	}
}
