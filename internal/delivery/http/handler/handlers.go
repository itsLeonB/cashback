package handler

import (
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/provider"
)

type Handlers struct {
	Auth              *AuthHandler
	Friendship        *FriendshipHandler
	FriendshipRequest *FriendshipRequestHandler
	Profile           *ProfileHandler
	TransferMethod    *TransferMethodHandler
	Debt              *DebtHandler
	GroupExpense      *GroupExpenseHandler
	ExpenseItem       *ExpenseItemHandler
	OtherFee          *OtherFeeHandler
	ExpenseBill       *ExpenseBillHandler
}

func ProvideHandlers(logger ezutil.Logger, services *provider.Services) *Handlers {
	return &Handlers{
		NewAuthHandler(services.Auth),
		NewFriendshipHandler(services.Friendship, services.FriendDetails),
		NewFriendshipRequestHandler(services.FriendshipRequest),
		NewProfileHandler(services.Profile),
		NewTransferMethodHandler(services.TransferMethod),
		NewDebtHandler(services.Debt),
		NewGroupExpenseHandler(services.GroupExpense),
		NewExpenseItemHandler(services.ExpenseItem),
		NewOtherFeeHandler(services.OtherFee),
		NewExpenseBillHandler(logger, services.ExpenseBill),
	}
}
