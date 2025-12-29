package service

import (
	"context"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	expenseV1 "github.com/itsLeonB/billsplittr-protos/gen/go/groupexpense/v1"
	expenseV2 "github.com/itsLeonB/billsplittr-protos/gen/go/groupexpense/v2"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain"
	"github.com/itsLeonB/orcashtrator/internal/domain/groupexpense"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/orcashtrator/internal/util"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
)

type groupExpenseServiceImpl struct {
	friendshipService  FriendshipService
	debtService        DebtService
	profileService     ProfileService
	groupExpenseClient groupexpense.GroupExpenseClient
	expenseClientV1    expenseV1.GroupExpenseServiceClient
	expenseClientV2    expenseV2.GroupExpenseServiceClient
	billSvc            ExpenseBillService
}

func NewGroupExpenseService(
	friendshipService FriendshipService,
	debtService DebtService,
	profileService ProfileService,
	groupExpenseClient groupexpense.GroupExpenseClient,
	expenseClientV1 expenseV1.GroupExpenseServiceClient,
	expenseClientV2 expenseV2.GroupExpenseServiceClient,
	billSvc ExpenseBillService,
) GroupExpenseService {
	return &groupExpenseServiceImpl{
		friendshipService,
		debtService,
		profileService,
		groupExpenseClient,
		expenseClientV1,
		expenseClientV2,
		billSvc,
	}
}

func (ges *groupExpenseServiceImpl) CreateDraft(ctx context.Context, request dto.NewGroupExpenseRequest) (dto.GroupExpenseResponse, error) {
	if err := ges.validateRequest(request); err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	// Default PayerProfileID to the user's profile ID if not provided
	// This is useful when the user is creating a group expense for themselves.
	if request.PayerProfileID == uuid.Nil {
		request.PayerProfileID = request.CreatorProfileID
	} else if request.PayerProfileID != request.CreatorProfileID {
		// Check if the payer is a friend of the user
		isFriend, _, err := ges.friendshipService.IsFriends(ctx, request.CreatorProfileID, request.PayerProfileID)
		if err != nil {
			return dto.GroupExpenseResponse{}, err
		}
		if !isFriend {
			return dto.GroupExpenseResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNotFriends)
		}
	}

	groupExpense := mapper.GroupExpenseRequestToEntity(request)
	groupExpense.CreatorProfileID = request.CreatorProfileID

	insertedGroupExpense, err := ges.groupExpenseClient.CreateDraft(ctx, groupExpense)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	profilesByID, err := ges.profileService.GetByIDs(ctx, insertedGroupExpense.ProfileIDs())
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	return mapper.GroupExpenseToResponse(insertedGroupExpense, request.CreatorProfileID, profilesByID, dto.ExpenseBillResponse{}), nil
}

func (ges *groupExpenseServiceImpl) GetAllCreated(ctx context.Context, userProfileID uuid.UUID) ([]dto.GroupExpenseResponse, error) {
	groupExpenses, err := ges.groupExpenseClient.GetAllCreated(ctx, userProfileID)
	if err != nil {
		return nil, err
	}

	profileIDs := make([]uuid.UUID, 0)
	for _, groupExpense := range groupExpenses {
		profileIDs = append(profileIDs, groupExpense.ProfileIDs()...)
	}

	profilesByID := make(map[uuid.UUID]dto.ProfileResponse, len(profileIDs))
	if len(profileIDs) > 0 {
		profilesByID, err = ges.profileService.GetByIDs(ctx, profileIDs)
		if err != nil {
			return nil, err
		}
	}

	mapFunc := func(groupExpense groupexpense.GroupExpense) dto.GroupExpenseResponse {
		return mapper.GroupExpenseToResponse(groupExpense, userProfileID, profilesByID, dto.ExpenseBillResponse{})
	}

	return ezutil.MapSlice(groupExpenses, mapFunc), nil
}

func (ges *groupExpenseServiceImpl) GetDetails(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error) {
	groupExpense, err := ges.groupExpenseClient.GetDetails(ctx, id)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	var eg errgroup.Group
	var billResponse dto.ExpenseBillResponse
	var profilesByID map[uuid.UUID]dto.ProfileResponse

	eg.Go(func() error {
		bill, err := ges.billSvc.MapToURL(ctx, groupExpense.Bill)
		billResponse = bill
		return err
	})

	eg.Go(func() error {
		profiles, err := ges.profileService.GetByIDs(ctx, groupExpense.ProfileIDs())
		profilesByID = profiles
		return err
	})

	if err = eg.Wait(); err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	expenseResponse := mapper.GroupExpenseToResponse(groupExpense, userProfileID, profilesByID, billResponse)
	return expenseResponse, nil
}

func (ges *groupExpenseServiceImpl) ConfirmDraft(ctx context.Context, id, userProfileID uuid.UUID) (dto.GroupExpenseResponse, error) {
	request := groupexpense.ConfirmDraftRequest{
		ID:        id,
		ProfileID: userProfileID,
	}

	groupExpense, err := ges.groupExpenseClient.ConfirmDraft(ctx, request)
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	if err = ges.debtService.ProcessConfirmedGroupExpense(ctx, groupExpense); err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	profilesByID, err := ges.profileService.GetByIDs(ctx, groupExpense.ProfileIDs())
	if err != nil {
		return dto.GroupExpenseResponse{}, err
	}

	return mapper.GroupExpenseToResponse(groupExpense, userProfileID, profilesByID, dto.ExpenseBillResponse{}), nil
}

func (ges *groupExpenseServiceImpl) CreateDraftV2(ctx context.Context, userProfileID uuid.UUID, description string) (dto.ExpenseResponseV2, error) {
	req := &expenseV2.CreateDraftRequest{
		CreatorProfileId: userProfileID.String(),
		Description:      description,
	}

	resp, err := ges.expenseClientV2.CreateDraft(ctx, req)
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	expense := resp.GetGroupExpense()
	if expense == nil {
		return dto.ExpenseResponseV2{}, eris.New("response is nil")
	}

	metadata, err := domain.FromAuditMetadataProto(expense.GetAuditMetadata())
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	creatorProfileID, err := ezutil.Parse[uuid.UUID](expense.CreatorProfileId)
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	payerProfileID, err := ezutil.Parse[uuid.UUID](expense.PayerProfileId)
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	profileIDs := mapset.NewSet(userProfileID, creatorProfileID, payerProfileID)
	profileMap, err := ges.profileService.GetByIDs(ctx, profileIDs.ToSlice())
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	status, err := mapper.FromExpenseStatusProto(expense.GetStatus())
	if err != nil {
		return dto.ExpenseResponseV2{}, err
	}

	return dto.ExpenseResponseV2{
		ID:               metadata.ID,
		CreatedAt:        metadata.CreatedAt,
		UpdatedAt:        metadata.UpdatedAt,
		DeletedAt:        metadata.DeletedAt,
		Creator:          mapper.ProfileResponseToParticipant(profileMap[creatorProfileID], userProfileID),
		Payer:            mapper.ProfileResponseToParticipant(profileMap[payerProfileID], userProfileID),
		TotalAmount:      decimal.Zero,
		ItemsTotalAmount: decimal.Zero,
		FeesTotalAmount:  decimal.Zero,
		Description:      expense.Description,
		Status:           status,
	}, nil
}

func (ges *groupExpenseServiceImpl) Delete(ctx context.Context, userProfileID, id uuid.UUID) error {
	req := &expenseV1.DeleteRequest{
		Id:        id.String(),
		ProfileId: userProfileID.String(),
	}
	_, err := ges.expenseClientV1.Delete(ctx, req)
	return err
}

func (ges *groupExpenseServiceImpl) SyncParticipants(ctx context.Context, req dto.ExpenseParticipantsRequest) error {
	participantSet := mapset.NewSet[uuid.UUID]()
	for _, pid := range req.ParticipantProfileIDs {
		participantSet.Add(pid)
	}
	if participantSet.Cardinality() != len(req.ParticipantProfileIDs) {
		return ungerr.UnprocessableEntityError("duplicate participant profile IDs given")
	}
	if !participantSet.Contains(req.PayerProfileID) {
		return ungerr.UnprocessableEntityError("payer profile ID must be one of the participant profile IDs")
	}

	for _, participantProfileID := range req.ParticipantProfileIDs {
		if participantProfileID == req.UserProfileID {
			continue
		}
		isFriends, _, err := ges.friendshipService.IsFriends(ctx, req.UserProfileID, participantProfileID)
		if err != nil {
			return err
		}
		if !isFriends {
			return ungerr.UnprocessableEntityError(appconstant.ErrNotFriends)
		}
	}

	if !participantSet.Contains(req.UserProfileID) {
		participantSet.Add(req.UserProfileID)
	}

	request := &expenseV1.SyncParticipantsRequest{
		ParticipantProfileIds: ezutil.MapSlice(participantSet.ToSlice(), util.ToString),
		PayerProfileId:        req.PayerProfileID.String(),
		UserProfileId:         req.UserProfileID.String(),
		GroupExpenseId:        req.GroupExpenseID.String(),
	}

	_, err := ges.expenseClientV1.SyncParticipants(ctx, request)
	return err
}

func (ges *groupExpenseServiceImpl) validateRequest(request dto.NewGroupExpenseRequest) error {
	if request.TotalAmount.LessThanOrEqual(decimal.Zero) {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountZero)
	}

	calculatedFeeTotal := decimal.Zero
	calculatedSubtotal := decimal.Zero
	for _, item := range request.Items {
		calculatedSubtotal = calculatedSubtotal.Add(item.Amount.Mul(decimal.NewFromInt(int64(item.Quantity))))
	}
	for _, fee := range request.OtherFees {
		calculatedFeeTotal = calculatedFeeTotal.Add(fee.Amount)
	}
	if calculatedFeeTotal.Add(calculatedSubtotal).Cmp(request.TotalAmount) != 0 {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountMismatched)
	}
	if calculatedSubtotal.Cmp(request.Subtotal) != 0 {
		return ungerr.UnprocessableEntityError(appconstant.ErrAmountMismatched)
	}

	return nil
}
