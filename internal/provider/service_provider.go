package provider

import (
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/config"
	"github.com/itsLeonB/orcashtrator/internal/service"
	"github.com/rotisserie/eris"
)

type Services struct {
	Auth              service.AuthService
	Profile           service.ProfileService
	Friendship        service.FriendshipService
	FriendshipRequest service.FriendshipRequestService
	TransferMethod    service.TransferMethodService
	Debt              service.DebtService
	FriendDetails     service.FriendDetailsService
	GroupExpense      service.GroupExpenseService
	ExpenseItem       service.ExpenseItemService
	OtherFee          service.OtherFeeService
	ExpenseBill       service.ExpenseBillService
}

func ProvideServices(
	clients *Clients,
	logger ezutil.Logger,
	cfg config.Config,
	queues *Queues,
) (*Services, error) {
	if clients == nil {
		return nil, eris.New("clients cannot be nil")
	}
	if logger == nil {
		return nil, eris.New("logger cannot be nil")
	}
	if queues == nil {
		return nil, eris.New("queue cannot be nil")
	}

	authService := service.NewAuthService(clients.Auth, cfg.RegisterVerificationUrl, cfg.ResetPasswordUrl)

	profileService := service.NewProfileService(clients.Profile)

	friendshipService := service.NewFriendshipService(
		clients.Friendship,
	)

	friendshipRequestService := service.NewFriendshipRequestService(clients.FriendshipRequest)

	transferMethodService := service.NewTransferMethodService(clients.TransferMethod)

	debtService := service.NewDebtService(
		clients.Debt,
		friendshipService,
	)

	friendDetailsService := service.NewFriendDetailsService(clients.Friendship, debtService)

	groupExpenseService := service.NewGroupExpenseService(
		friendshipService,
		debtService,
		profileService,
		clients.GroupExpense,
	)

	expenseItemSvc := service.NewExpenseItemService(
		profileService,
		clients.ExpenseItem,
	)

	otherFeeSvc := service.NewOtherFeeService(
		profileService,
		clients.OtherFee,
	)

	expenseBillService := service.NewExpenseBillService(
		logger,
		friendshipService,
		profileService,
		clients.ExpenseBill,
		clients.ImageUpload,
		cfg.BucketNameExpenseBill,
		queues.ExpenseBillUploaded,
	)

	return &Services{
		authService,
		profileService,
		friendshipService,
		friendshipRequestService,
		transferMethodService,
		debtService,
		friendDetailsService,
		groupExpenseService,
		expenseItemSvc,
		otherFeeSvc,
		expenseBillService,
	}, nil
}
