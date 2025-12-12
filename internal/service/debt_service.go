package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain/debt"
	"github.com/itsLeonB/orcashtrator/internal/domain/groupexpense"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

type debtServiceImpl struct {
	debtClient        debt.DebtClient
	friendshipService FriendshipService
	profileService    ProfileService
}

func NewDebtService(
	debtClient debt.DebtClient,
	friendshipService FriendshipService,
	profileService ProfileService,
) DebtService {
	return &debtServiceImpl{
		debtClient,
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

	request := debt.RecordNewTransactionRequest{
		UserProfileID:    req.UserProfileID,
		FriendProfileID:  req.FriendProfileID,
		Action:           req.Action,
		Amount:           req.Amount,
		TransferMethodID: req.TransferMethodID,
		Description:      req.Description,
	}

	transaction, err := ds.debtClient.RecordNewTransaction(ctx, request)
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	return mapper.DebtTransactionToResponse(transaction), nil
}

func (ds *debtServiceImpl) GetTransactions(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	transactions, err := ds.debtClient.GetTransactions(ctx, profileID)
	if err != nil {
		return nil, eris.Wrap(err, appconstant.ErrServiceClient)
	}

	return ezutil.MapSlice(transactions, mapper.DebtTransactionToResponse), nil
}

func (ds *debtServiceImpl) ProcessConfirmedGroupExpense(ctx context.Context, groupExpense groupexpense.GroupExpense) error {
	if !groupExpense.IsConfirmed {
		return ungerr.UnprocessableEntityError("group expense is not confirmed")
	}

	if len(groupExpense.Participants) < 1 {
		return ungerr.UnprocessableEntityError("no participants to process")
	}

	participants, err := ezutil.MapSliceWithError(groupExpense.Participants, mapper.ExpenseParticipantToData)
	if err != nil {
		return err
	}

	request := debt.GroupExpenseData{
		ID:               groupExpense.ID,
		PayerProfileID:   groupExpense.PayerProfileID,
		CreatorProfileID: groupExpense.CreatorProfileID,
		Description:      groupExpense.Description,
		Participants:     participants,
	}

	return ds.debtClient.ProcessConfirmedGroupExpense(ctx, request)
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

	allTransactions := make([]debt.Transaction, 0)
	for _, profileID := range associatedProfileIDs {
		request := debt.GetAllByProfileIDsRequest{
			UserProfileID:   userProfileID,
			FriendProfileID: profileID,
		}

		transactions, err := ds.debtClient.GetAllByProfileIDs(ctx, request)
		if err != nil {
			return nil, eris.Wrap(err, appconstant.ErrServiceClient)
		}

		allTransactions = append(allTransactions, transactions...)
	}

	return ezutil.MapSlice(allTransactions, mapper.DebtTransactionToResponse), nil
}
