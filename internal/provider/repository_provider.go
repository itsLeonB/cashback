package provider

import (
	adapters "github.com/itsLeonB/cashback/internal/adapters/repository"
	monetizationAdapter "github.com/itsLeonB/cashback/internal/adapters/repository/monetization"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/monetization"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	monetizationRepo "github.com/itsLeonB/cashback/internal/domain/repository/monetization"
	"github.com/itsLeonB/go-crud"
	"gorm.io/gorm"
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
	Session            crud.Repository[users.Session]
	RefreshToken       crud.Repository[users.RefreshToken]

	// Debts
	DebtTransaction       repository.DebtTransactionRepository
	TransferMethod        repository.TransferMethodRepository
	ProfileTransferMethod crud.Repository[debts.ProfileTransferMethod]

	// Expenses
	GroupExpense repository.GroupExpenseRepository
	ExpenseItem  repository.ExpenseItemRepository
	OtherFee     repository.OtherFeeRepository
	ExpenseBill  repository.ExpenseBillRepository

	// Monetization
	Plan         crud.Repository[monetization.Plan]
	PlanVersion  monetizationRepo.PlanVersionRepository
	Subscription monetizationRepo.SubscriptionRepository
	Payment      crud.Repository[monetization.Payment]

	// Infra
	Notification     repository.NotificationRepository
	PushSubscription repository.PushSubscriptionRepository
}

func ProvideRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Transactor: crud.NewTransactor(db),

		User:               crud.NewRepository[users.User](db),
		Profile:            adapters.NewProfileRepository(db),
		Friendship:         adapters.NewFriendshipRepository(db),
		RelatedProfile:     crud.NewRepository[users.RelatedProfile](db),
		PasswordResetToken: crud.NewRepository[users.PasswordResetToken](db),
		OAuthAccount:       crud.NewRepository[users.OAuthAccount](db),
		FriendshipRequest:  crud.NewRepository[users.FriendshipRequest](db),
		Session:            crud.NewRepository[users.Session](db),
		RefreshToken:       crud.NewRepository[users.RefreshToken](db),

		DebtTransaction:       adapters.NewDebtTransactionRepository(db),
		TransferMethod:        adapters.NewTransferMethodRepository(db),
		ProfileTransferMethod: crud.NewRepository[debts.ProfileTransferMethod](db),

		GroupExpense: adapters.NewGroupExpenseRepository(db),
		ExpenseItem:  adapters.NewExpenseItemRepository(db),
		OtherFee:     adapters.NewOtherFeeRepository(db),
		ExpenseBill:  adapters.NewExpenseBillRepository(db),

		Plan:         crud.NewRepository[monetization.Plan](db),
		PlanVersion:  monetizationAdapter.NewPlanVersionRepository(db),
		Subscription: monetizationAdapter.NewSubscriptionRepository(db),
		Payment:      crud.NewRepository[monetization.Payment](db),

		Notification:     adapters.NewNotificationRepository(db),
		PushSubscription: adapters.NewPushSubscriptionRepository(db),
	}
}
