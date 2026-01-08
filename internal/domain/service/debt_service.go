package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/cashback/internal/domain/service/debt"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type debtServiceImpl struct {
	debtCalculatorStrategy    map[debts.DebtTransactionAction]debt.DebtCalculator
	debtTransactionRepository repository.DebtTransactionRepository
	transferMethodService     TransferMethodService
	friendshipService         FriendshipService
	profileService            ProfileService
}

func NewDebtService(
	debtCalculatorStrategy map[debts.DebtTransactionAction]debt.DebtCalculator,
	debtTransactionRepository repository.DebtTransactionRepository,
	transferMethodService TransferMethodService,
	friendshipService FriendshipService,
	profileService ProfileService,
) DebtService {
	return &debtServiceImpl{
		debtCalculatorStrategy,
		debtTransactionRepository,
		transferMethodService,
		friendshipService,
		profileService,
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

	calculator, err := ds.selectCalculator(request.Action)
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	insertedDebt, err := ds.debtTransactionRepository.Insert(ctx, calculator.MapRequestToEntity(request))
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	insertedDebt.TransferMethod = transferMethod
	return calculator.MapEntityToResponse(insertedDebt), nil
}

func (ds *debtServiceImpl) selectCalculator(action debts.DebtTransactionAction) (debt.DebtCalculator, error) {
	calculator, ok := ds.debtCalculatorStrategy[action]
	if !ok {
		return nil, ungerr.Unknownf("unsupported debt calculator action: %s", action)
	}

	return calculator, nil
}

func (ds *debtServiceImpl) GetTransactions(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	transactions, err := ds.debtTransactionRepository.FindAllByUserProfileID(ctx, profileID)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(transactions, mapper.DebtTransactionSimpleMapper(profileID)), nil
}

func (ds *debtServiceImpl) ProcessConfirmedGroupExpense(ctx context.Context, groupExpense expenses.GroupExpense) error {
	if groupExpense.Status != expenses.ConfirmedExpense {
		return ungerr.UnprocessableEntityError("group expense is not confirmed")
	}
	if len(groupExpense.Participants) < 1 {
		return ungerr.UnprocessableEntityError("no participants to process")
	}

	transferMethod, err := ds.transferMethodService.GetByName(ctx, debts.GroupExpenseTransferMethod)
	if err != nil {
		return err
	}

	debtTransactions := mapper.GroupExpenseToDebtTransactions(groupExpense, transferMethod.ID)

	_, err = ds.debtTransactionRepository.InsertMany(ctx, debtTransactions)
	return err
}

func (ds *debtServiceImpl) GetAllByProfileIDs(ctx context.Context, userProfileID, friendProfileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	friendProfile, err := ds.profileService.GetByID(ctx, friendProfileID)
	if err != nil {
		return nil, err
	}

	associatedProfileIDs := []uuid.UUID{friendProfileID}
	if friendProfile.IsAnonymous {
		if friendProfile.RealProfileID != uuid.Nil {
			associatedProfileIDs = append(associatedProfileIDs, friendProfile.RealProfileID)
		}
	} else {
		associatedProfileIDs = append(associatedProfileIDs, friendProfile.AssociatedAnonProfileIDs...)
	}

	allTransactions := make([]debts.DebtTransaction, 0)
	for _, profileID := range associatedProfileIDs {
		transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, userProfileID, profileID)
		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, transactions...)
	}

	return ezutil.MapSlice(allTransactions, mapper.DebtTransactionSimpleMapper(userProfileID)), nil
}
