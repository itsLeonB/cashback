package provider

import (
	adapters "github.com/itsLeonB/cashback/internal/adapters/repository"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/go-crud"
)

type Repositories struct {
	Transactor crud.Transactor

	// Users
	User               crud.Repository[users.User]
	Profile            repository.ProfileRepository
	Friendship         repository.FriendshipRepository
	RelatedProfile     crud.Repository[users.RelatedProfile]
	PasswordResetToken crud.Repository[users.PasswordResetToken]
	OAuthAccount       crud.Repository[users.OAuthAccount]
	FriendshipRequest  crud.Repository[users.FriendshipRequest]

	// Debts
	DebtTransaction       repository.DebtTransactionRepository
	TransferMethod        repository.TransferMethodRepository
	ProfileTransferMethod crud.Repository[debts.ProfileTransferMethod]

	// Expenses
	GroupExpense repository.GroupExpenseRepository
	ExpenseItem  repository.ExpenseItemRepository
	OtherFee     repository.OtherFeeRepository
	ExpenseBill  crud.Repository[expenses.ExpenseBill]

	// Infra
	Notification     repository.NotificationRepository
	PushSubscription repository.PushSubscriptionRepository
}

func ProvideRepositories(dataSource *DataSources) *Repositories {
	return &Repositories{
		Transactor:            crud.NewTransactor(dataSource.Gorm),
		User:                  crud.NewRepository[users.User](dataSource.Gorm),
		Profile:               adapters.NewProfileRepository(dataSource.Gorm),
		Friendship:            adapters.NewFriendshipRepository(dataSource.Gorm),
		RelatedProfile:        crud.NewRepository[users.RelatedProfile](dataSource.Gorm),
		PasswordResetToken:    crud.NewRepository[users.PasswordResetToken](dataSource.Gorm),
		OAuthAccount:          crud.NewRepository[users.OAuthAccount](dataSource.Gorm),
		FriendshipRequest:     crud.NewRepository[users.FriendshipRequest](dataSource.Gorm),
		DebtTransaction:       adapters.NewDebtTransactionRepository(dataSource.Gorm),
		TransferMethod:        adapters.NewTransferMethodRepository(dataSource.Gorm),
		ProfileTransferMethod: crud.NewRepository[debts.ProfileTransferMethod](dataSource.Gorm),
		GroupExpense:          adapters.NewGroupExpenseRepository(dataSource.Gorm),
		ExpenseItem:           adapters.NewExpenseItemRepository(dataSource.Gorm),
		OtherFee:              adapters.NewOtherFeeRepository(dataSource.Gorm),
		ExpenseBill:           crud.NewRepository[expenses.ExpenseBill](dataSource.Gorm),
		Notification:          adapters.NewNotificationRepository(dataSource.Gorm),
		PushSubscription:      adapters.NewPushSubscriptionRepository(dataSource.Gorm),
	}
}
