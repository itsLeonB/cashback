package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
)

type DebtTransactionRepository interface {
	crud.Repository[debts.DebtTransaction]
	FindAllByMultipleProfileIDs(ctx context.Context, userProfileIDs, friendProfileIDs []uuid.UUID) ([]debts.DebtTransaction, error)
	FindAllByUserProfileID(ctx context.Context, userProfileID uuid.UUID) ([]debts.DebtTransaction, error)
}

type GroupExpenseRepository interface {
	crud.Repository[expenses.GroupExpense]
	SyncParticipants(ctx context.Context, groupExpenseID uuid.UUID, participants []expenses.ExpenseParticipant) error
	DeleteItemParticipants(ctx context.Context, expenseID uuid.UUID, newParticipantProfileIDs []uuid.UUID) error
}

type ExpenseItemRepository interface {
	crud.Repository[expenses.ExpenseItem]
	SyncParticipants(ctx context.Context, expenseItemID uuid.UUID, participants []expenses.ItemParticipant) error
}

type OtherFeeRepository interface {
	crud.Repository[expenses.OtherFee]
	SyncParticipants(ctx context.Context, feeID uuid.UUID, participants []expenses.FeeParticipant) error
}

type ProfileRepository interface {
	crud.Repository[users.UserProfile]
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.UserProfile, error)
	SearchByName(ctx context.Context, query string, limit int) ([]users.UserProfile, error)
}

type FriendshipRepository interface {
	crud.Repository[users.Friendship]
	Insert(ctx context.Context, friendship users.Friendship) (users.Friendship, error)
	FindAllBySpec(ctx context.Context, spec users.FriendshipSpecification) ([]users.Friendship, error)
	FindFirstBySpec(ctx context.Context, spec users.FriendshipSpecification) (users.Friendship, error)
	FindByProfileIDs(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error)
}

type TransferMethodRepository interface {
	crud.Repository[debts.TransferMethod]
	GetAllByParentFilter(ctx context.Context, filter debts.ParentFilter, profileID uuid.UUID) ([]debts.TransferMethod, error)
}
