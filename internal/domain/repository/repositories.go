package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/go-crud"
	"github.com/shopspring/decimal"
)

type DebtTransactionRepository interface {
	crud.Repository[debts.DebtTransaction]
	FindAllByMultipleProfileIDs(ctx context.Context, userProfileIDs, friendProfileIDs []uuid.UUID) ([]debts.DebtTransaction, error)
	FindAllByProfileIDs(ctx context.Context, profileIDs []uuid.UUID, limit int, debtsOnly bool) ([]debts.DebtTransaction, error)
	// RepointProfile repoints every debt transaction referencing anonProfileID (as lender or
	// borrower) onto realProfileID. No unique constraint on these columns, so this is a plain update.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}

type GroupExpenseRepository interface {
	crud.Repository[expenses.GroupExpense]
	SyncParticipants(ctx context.Context, groupExpenseID uuid.UUID, participants []expenses.ExpenseParticipant) error
	DeleteItemParticipants(ctx context.Context, expenseID uuid.UUID, newParticipantProfileIDs []uuid.UUID) error
	FindAllByOwnership(ctx context.Context, profileID uuid.UUID, ownership expenses.ExpenseOwnership, status expenses.ExpenseStatus, limit int) ([]expenses.GroupExpense, error)
	FindRecentByProfileID(ctx context.Context, profileID uuid.UUID, limit int) ([]expenses.GroupExpense, error)
	// RepointProfile repoints group_expenses.payer/creator and group_expense_participants
	// referencing anonProfileID onto realProfileID. If realProfileID is already a participant
	// in the same group expense as anonProfileID, their share amounts are summed and the
	// anonymous row is dropped instead of violating the unique(group_expense_id, profile) constraint.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}

type ExpenseItemRepository interface {
	crud.Repository[expenses.ExpenseItem]
	SyncParticipants(ctx context.Context, expenseItemID uuid.UUID, participants []expenses.ItemParticipant) error
	// RepointParticipants merges group_expense_item_participants rows referencing anonProfileID
	// onto realProfileID, summing weight/allocated_amount on collision (see GroupExpenseRepository.RepointProfile).
	RepointParticipants(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}

type OtherFeeRepository interface {
	crud.Repository[expenses.OtherFee]
	SyncParticipants(ctx context.Context, feeID uuid.UUID, participants []expenses.FeeParticipant) error
	// RepointParticipants merges group_expense_other_fee_participants rows referencing
	// anonProfileID onto realProfileID, summing share_amount on collision.
	RepointParticipants(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}

type ProfileRepository interface {
	crud.Repository[users.UserProfile]
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]users.UserProfile, error)
	FindRealProfiles(ctx context.Context) ([]users.UserProfile, error)
	SearchByName(ctx context.Context, query string, limit int) ([]users.UserProfile, error)
}

type FriendshipRepository interface {
	crud.Repository[users.Friendship]
	Insert(ctx context.Context, friendship users.Friendship) (users.Friendship, error)
	FindAllBySpec(ctx context.Context, spec users.FriendshipSpecification) ([]users.Friendship, error)
	FindFirstBySpec(ctx context.Context, spec users.FriendshipSpecification) (users.Friendship, error)
	FindByProfileIDs(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error)
	// RepointFriendships repoints every friendship row involving anonProfileID onto
	// realProfileID. If realProfileID already has a friendship with the same counterparty, the
	// anonymous row is dropped instead of violating the unique(profile_id1, profile_id2) constraint.
	RepointFriendships(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
	// FindByProfileIDsForUpdate is FindByProfileIDs with SELECT ... FOR UPDATE, so a balance
	// recalculation for this pair serializes against a concurrent one instead of both reading a
	// stale transaction set and racing to overwrite friendship_balances with a stale value.
	FindByProfileIDsForUpdate(ctx context.Context, profileID1, profileID2 uuid.UUID) (users.Friendship, error)
	// FindAllByProfileIDForUpdate is the bulk equivalent of FindByProfileIDsForUpdate, locking
	// every friendship row profileID is party to (used by whole-profile balance recalculation).
	FindAllByProfileIDForUpdate(ctx context.Context, profileID uuid.UUID) ([]users.Friendship, error)
}

// FriendshipBalanceRow is a join-result shape (friendship_balances x friendships), not a
// domain entity - just enough to resolve sign/counterparty per row without a second query.
type FriendshipBalanceRow struct {
	ProfileID1 uuid.UUID
	ProfileID2 uuid.UUID
	Currency   string
	NetBalance decimal.Decimal
}

type FriendshipBalanceRepository interface {
	crud.Repository[users.FriendshipBalance]
	// UpsertMany always overwrites (friendship_id, currency) with a freshly computed value,
	// never an increment, so it self-corrects regardless of call order.
	UpsertMany(ctx context.Context, balances []users.FriendshipBalance) error
	FindAllByFriendshipID(ctx context.Context, friendshipID uuid.UUID) ([]users.FriendshipBalance, error)
	// FindAllByProfileID returns every non-zero balance row for a friendship profileID is party
	// to, joined against friendships to resolve the pair's profile IDs in one query.
	FindAllByProfileID(ctx context.Context, profileID uuid.UUID) ([]FriendshipBalanceRow, error)
	// DeleteZeroBalances removes rows that have settled to exactly 0 for the given friendships.
	// Called after UpsertMany, which must still write the 0 first to correct any stale nonzero
	// row - this only sweeps away what's left at 0, keeping the table free of dead weight.
	DeleteZeroBalances(ctx context.Context, friendshipIDs []uuid.UUID) error
}

type TransferMethodRepository interface {
	crud.Repository[debts.TransferMethod]
	GetAllByParentFilter(ctx context.Context, filter debts.ParentFilter, profileID uuid.UUID) ([]debts.TransferMethod, error)
}

type FriendshipRequestRepository interface {
	crud.Repository[users.FriendshipRequest]
	// RepointProfile repoints every friendship request referencing anonProfileID (as sender or
	// recipient) onto realProfileID. No unique constraint on these columns, so this is a plain update.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}

type ProfileTransferMethodRepository interface {
	crud.Repository[debts.ProfileTransferMethod]
	// RepointProfile repoints every transfer method referencing anonProfileID onto
	// realProfileID. No unique constraint on this column, so this is a plain update.
	RepointProfile(ctx context.Context, anonProfileID, realProfileID uuid.UUID) error
}
