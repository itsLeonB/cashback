package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/ungerr"
)

type expenseItemServiceImpl struct {
	profileService    ProfileService
	expenseItemClient expenseitem.ExpenseItemClient
}

func NewExpenseItemService(
	profileService ProfileService,
	expenseItemClient expenseitem.ExpenseItemClient,
) ExpenseItemService {
	return &expenseItemServiceImpl{
		profileService,
		expenseItemClient,
	}
}

func (ges *expenseItemServiceImpl) Add(ctx context.Context, req dto.NewExpenseItemRequest) (dto.ExpenseItemResponse, error) {
	if !req.Amount.IsPositive() {
		return dto.ExpenseItemResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	request := expenseitem.AddRequest{
		ProfileID:       req.UserProfileID,
		GroupExpenseID:  req.GroupExpenseID,
		ExpenseItemData: mapper.NewExpenseItemRequestToData(req),
	}

	expenseItem, err := ges.expenseItemClient.Add(ctx, request)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	profileIDs := []uuid.UUID{req.UserProfileID}
	profileIDs = append(profileIDs, expenseItem.ProfileIDs()...)
	profilesByID, err := ges.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	return mapper.ExpenseItemToResponse(expenseItem, req.UserProfileID, profilesByID), nil
}

func (ges *expenseItemServiceImpl) GetDetails(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) (dto.ExpenseItemResponse, error) {
	request := expenseitem.GetDetailsRequest{
		ID:             expenseItemID,
		GroupExpenseID: groupExpenseID,
	}

	expenseItem, err := ges.expenseItemClient.GetDetails(ctx, request)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	profileIDs := []uuid.UUID{userProfileID}
	profileIDs = append(profileIDs, expenseItem.ProfileIDs()...)
	profilesByID, err := ges.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	return mapper.ExpenseItemToResponse(expenseItem, userProfileID, profilesByID), nil
}

func (ges *expenseItemServiceImpl) Update(ctx context.Context, req dto.UpdateExpenseItemRequest) (dto.ExpenseItemResponse, error) {
	if !req.Amount.IsPositive() {
		return dto.ExpenseItemResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	request := expenseitem.UpdateRequest{
		ProfileID:       req.UserProfileID,
		ID:              req.ID,
		GroupExpenseID:  req.GroupExpenseID,
		ExpenseItemData: mapper.UpdateExpenseItemRequestToData(req),
	}

	expenseItem, err := ges.expenseItemClient.Update(ctx, request)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	profileIDs := []uuid.UUID{req.UserProfileID}
	profileIDs = append(profileIDs, expenseItem.ProfileIDs()...)
	profilesByID, err := ges.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return dto.ExpenseItemResponse{}, err
	}

	return mapper.ExpenseItemToResponse(expenseItem, req.UserProfileID, profilesByID), nil
}

func (ges *expenseItemServiceImpl) Remove(ctx context.Context, groupExpenseID, expenseItemID, userProfileID uuid.UUID) error {
	request := expenseitem.RemoveRequest{
		ProfileID:      userProfileID,
		ID:             expenseItemID,
		GroupExpenseID: groupExpenseID,
	}

	return ges.expenseItemClient.Remove(ctx, request)
}
