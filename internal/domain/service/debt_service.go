package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type debtServiceImpl struct {
	debtTransactionRepository repository.DebtTransactionRepository
	transferMethodService     TransferMethodService
	friendshipService         FriendshipService
	profileService            ProfileService
	expenseService            GroupExpenseService
}

func NewDebtService(
	debtTransactionRepository repository.DebtTransactionRepository,
	transferMethodService TransferMethodService,
	friendshipService FriendshipService,
	profileService ProfileService,
	expenseService GroupExpenseService,
) DebtService {
	return &debtServiceImpl{
		debtTransactionRepository,
		transferMethodService,
		friendshipService,
		profileService,
		expenseService,
	}
}

func (ds *debtServiceImpl) RecordNewTransaction(ctx context.Context, req dto.NewDebtTransactionRequest) (dto.DebtTransactionResponse, error) {
	if req.Amount.Compare(decimal.Zero) < 1 {
		return dto.DebtTransactionResponse{}, ungerr.ValidationError("amount must be greater than 0")
	}
	if req.UserProfileID == req.FriendProfileID {
		return dto.DebtTransactionResponse{}, ungerr.UnprocessableEntityError("cannot do self transactions")
	}
	isFriends, _, err := ds.friendshipService.IsFriends(ctx, req.UserProfileID, req.FriendProfileID)
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}
	if !isFriends {
		return dto.DebtTransactionResponse{}, ungerr.UnprocessableEntityError("both profiles are not friends")
	}

	return ds.recordNew(ctx, req)
}

func (ds *debtServiceImpl) recordNew(ctx context.Context, request dto.NewDebtTransactionRequest) (dto.DebtTransactionResponse, error) {
	if !request.Amount.IsPositive() {
		return dto.DebtTransactionResponse{}, ungerr.ValidationError("amount must be greater than 0")
	}

	transferMethod, err := ds.transferMethodService.GetByID(ctx, request.TransferMethodID)
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	lenderID, borrowerID := request.UserProfileID, request.FriendProfileID
	if request.Direction == dto.IncomingDebt {
		lenderID, borrowerID = request.FriendProfileID, request.UserProfileID
	}

	insertedDebt, err := ds.debtTransactionRepository.Insert(ctx, debts.DebtTransaction{
		LenderProfileID:   lenderID,
		BorrowerProfileID: borrowerID,
		Amount:            request.Amount,
		TransferMethodID:  request.TransferMethodID,
		Description:       request.Description,
	})
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	insertedDebt.TransferMethod = transferMethod
	return mapper.DebtTransactionToResponse(request.UserProfileID, insertedDebt), nil
}

func (ds *debtServiceImpl) GetTransactions(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	profile, err := ds.profileService.GetByID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	profileIDs := ds.getAssociatedIDs(profile)

	transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, profileIDs)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(transactions, mapper.DebtTransactionSimpleMapper(profileID)), nil
}

func (ds *debtServiceImpl) GetTransactionSummary(ctx context.Context, profileID uuid.UUID) (dto.FriendBalance, error) {
	profile, err := ds.profileService.GetByID(ctx, profileID)
	if err != nil {
		return dto.FriendBalance{}, err
	}

	profileIDs := ds.getAssociatedIDs(profile)

	transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, profileIDs)
	if err != nil {
		return dto.FriendBalance{}, err
	}

	return mapper.MapToFriendBalanceSummary(transactions, profileIDs), nil
}

func (ds *debtServiceImpl) ProcessConfirmedGroupExpense(ctx context.Context, msg message.ExpenseConfirmed) error {
	groupExpense, err := ds.expenseService.GetByID(ctx, msg.ID)
	if err != nil {
		return err
	}

	if groupExpense.Status != expenses.ConfirmedExpense {
		return ungerr.Unknown("group expense is not confirmed")
	}
	if len(groupExpense.Participants) < 1 {
		return ungerr.Unknown("no participants to process")
	}

	transferMethod, err := ds.transferMethodService.GetByName(ctx, debts.GroupExpenseTransferMethod)
	if err != nil {
		return err
	}

	debtTransactions := mapper.GroupExpenseToDebtTransactions(groupExpense, transferMethod.ID)

	_, err = ds.debtTransactionRepository.InsertMany(ctx, debtTransactions)
	return err
}

func (ds *debtServiceImpl) GetAllByProfileIDs(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]debts.DebtTransaction, []uuid.UUID, error) {
	profiles, err := ds.profileService.GetByIDs(ctx, []uuid.UUID{userProfileID, friendProfileID})
	if err != nil {
		return nil, nil, err
	}

	userIDs := ds.getAssociatedIDs(profiles[userProfileID])
	friendIDs := ds.getAssociatedIDs(profiles[friendProfileID])

	transactions, err := ds.debtTransactionRepository.FindAllByMultipleProfileIDs(ctx, userIDs, friendIDs)
	return transactions, userIDs, err
}

func (ds *debtServiceImpl) getAssociatedIDs(profile dto.ProfileResponse) []uuid.UUID {
	ids := []uuid.UUID{profile.ID}
	if profile.IsAnonymous {
		if profile.RealProfileID != uuid.Nil {
			ids = append(ids, profile.RealProfileID)
		}
	} else {
		ids = append(ids, profile.AssociatedAnonProfileIDs...)
	}
	return ids
}
