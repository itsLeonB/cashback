package service

import (
	"context"
	"encoding/json"
	"fmt"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
	"gorm.io/datatypes"
)

type debtServiceImpl struct {
	debtTransactionRepository repository.DebtTransactionRepository
	transferMethodService     TransferMethodService
	friendshipService         FriendshipService
	profileService            ProfileService
	expenseService            GroupExpenseService
	taskQueue                 queue.TaskQueue
}

func NewDebtService(
	debtTransactionRepository repository.DebtTransactionRepository,
	transferMethodService TransferMethodService,
	friendshipService FriendshipService,
	profileService ProfileService,
	expenseService GroupExpenseService,
	taskQueue queue.TaskQueue,
) DebtService {
	return &debtServiceImpl{
		debtTransactionRepository,
		transferMethodService,
		friendshipService,
		profileService,
		expenseService,
		taskQueue,
	}
}

func (ds *debtServiceImpl) RecordNewTransaction(ctx context.Context, req dto.NewDebtTransactionRequest) (dto.DebtTransactionResponse, error) {
	if !req.Amount.IsPositive() {
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

	transferMethod, err := ds.transferMethodService.GetByID(ctx, req.TransferMethodID)
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	lenderID, borrowerID := req.UserProfileID, req.FriendProfileID
	if req.Direction == dto.IncomingDebt {
		lenderID, borrowerID = req.FriendProfileID, req.UserProfileID
	}

	insertedDebt, err := ds.debtTransactionRepository.Insert(ctx, debts.DebtTransaction{
		LenderProfileID:   lenderID,
		BorrowerProfileID: borrowerID,
		Amount:            req.Amount,
		TransferMethodID:  req.TransferMethodID,
		Description:       req.Description,
	})
	if err != nil {
		return dto.DebtTransactionResponse{}, err
	}

	go func() {
		msg := message.DebtCreated{
			ID:               insertedDebt.ID,
			CreatorProfileID: req.UserProfileID,
		}
		if e := ds.taskQueue.Enqueue(context.Background(), msg); e != nil {
			logger.Errorf("error enqueuing %s: %v", msg.Type(), e)
		}
	}()

	insertedDebt.TransferMethod = transferMethod
	return mapper.DebtTransactionToResponse(req.UserProfileID, insertedDebt, make(map[uuid.UUID]dto.ProfileResponse)), nil
}

func (ds *debtServiceImpl) GetTransactions(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	profileIDs, err := ds.profileService.GetAssociatedIDs(ctx, profileID)
	if err != nil {
		return nil, err
	}

	transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, profileIDs, -1, false)
	if err != nil {
		return nil, err
	}

	trxProfileIDs := mapset.NewSet[uuid.UUID]()
	for _, transaction := range transactions {
		trxProfileIDs.Add(transaction.LenderProfileID)
		trxProfileIDs.Add(transaction.BorrowerProfileID)
	}

	profilesByID, err := ds.profileService.GetByIDs(ctx, trxProfileIDs.ToSlice())
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(transactions, mapper.DebtTransactionSimpleMapper(profileID, profilesByID)), nil
}

func (ds *debtServiceImpl) GetTransactionSummary(ctx context.Context, profileID uuid.UUID) (dto.FriendBalance, error) {
	profileIDs, err := ds.profileService.GetAssociatedIDs(ctx, profileID)
	if err != nil {
		return dto.FriendBalance{}, err
	}

	transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, profileIDs, -1, false)
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

func (ds *debtServiceImpl) GetRecent(ctx context.Context, profileID uuid.UUID) ([]dto.DebtTransactionResponse, error) {
	profileIDs, err := ds.profileService.GetAssociatedIDs(ctx, profileID)
	if err != nil {
		return nil, err
	}

	transactions, err := ds.debtTransactionRepository.FindAllByProfileIDs(ctx, profileIDs, 5, true)
	if err != nil {
		return nil, err
	}

	trxProfileIDs := mapset.NewSet[uuid.UUID]()
	for _, transaction := range transactions {
		trxProfileIDs.Add(transaction.LenderProfileID)
		trxProfileIDs.Add(transaction.BorrowerProfileID)
	}

	profilesByID, err := ds.profileService.GetByIDs(ctx, trxProfileIDs.ToSlice())
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(transactions, mapper.DebtTransactionSimpleMapper(profileID, profilesByID)), nil
}

func (ds *debtServiceImpl) ConstructNotification(ctx context.Context, msg message.DebtCreated) (entity.Notification, error) {
	spec := crud.Specification[debts.DebtTransaction]{}
	spec.Model.ID = msg.ID
	trx, err := ds.debtTransactionRepository.FindFirst(ctx, spec)
	if err != nil {
		return entity.Notification{}, err
	}
	if trx.IsZero() {
		return entity.Notification{}, ungerr.NotFoundError(fmt.Sprintf("debt transaction with ID: %s is not found", msg.ID))
	}

	toNotifyProfileID := trx.LenderProfileID
	if trx.LenderProfileID == msg.CreatorProfileID {
		toNotifyProfileID = trx.BorrowerProfileID
	}

	friendship, err := ds.friendshipService.GetByProfileIDs(ctx, trx.LenderProfileID, trx.BorrowerProfileID)
	if err != nil {
		return entity.Notification{}, err
	}

	otherParty := friendship.Profile2
	if toNotifyProfileID == friendship.ProfileID2 {
		otherParty = friendship.Profile1
	}

	msgMeta := message.DebtCreatedMetadata{
		FriendshipID: friendship.ID,
		FriendName:   otherParty.Name,
	}

	metadata, err := json.Marshal(msgMeta)
	if err != nil {
		return entity.Notification{}, err
	}

	return entity.Notification{
		ProfileID:  toNotifyProfileID,
		Type:       msg.Type(),
		EntityType: "debt-transaction",
		EntityID:   msg.ID,
		Metadata:   datatypes.JSON(metadata),
	}, nil
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
