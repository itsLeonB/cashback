package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/ungerr"
)

type otherFeeServiceImpl struct {
	profileService ProfileService
	otherFeeClient otherfee.OtherFeeClient
}

func NewOtherFeeService(
	profileService ProfileService,
	otherFeeClient otherfee.OtherFeeClient,
) OtherFeeService {
	return &otherFeeServiceImpl{
		profileService,
		otherFeeClient,
	}
}

func (ofs *otherFeeServiceImpl) Add(ctx context.Context, req dto.NewOtherFeeRequest) (dto.OtherFeeResponse, error) {
	if !req.Amount.IsPositive() {
		return dto.OtherFeeResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	request := otherfee.AddRequest{
		ProfileID:      req.UserProfileID,
		GroupExpenseID: req.GroupExpenseID,
		OtherFeeData: otherfee.OtherFeeData{
			Name:              req.Name,
			Amount:            req.Amount,
			CalculationMethod: req.CalculationMethod,
		},
	}

	otherFee, err := ofs.otherFeeClient.Add(ctx, request)
	if err != nil {
		return dto.OtherFeeResponse{}, err
	}

	profileIDs := []uuid.UUID{req.UserProfileID}
	profileIDs = append(profileIDs, otherFee.ProfileIDs()...)
	profilesByID, err := ofs.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return dto.OtherFeeResponse{}, err
	}

	return mapper.OtherFeeToResponse(otherFee, req.UserProfileID, profilesByID), nil
}

func (ofs *otherFeeServiceImpl) Update(ctx context.Context, req dto.UpdateOtherFeeRequest) (dto.OtherFeeResponse, error) {
	if !req.Amount.IsPositive() {
		return dto.OtherFeeResponse{}, ungerr.UnprocessableEntityError(appconstant.ErrNonPositiveAmount)
	}

	request := otherfee.UpdateRequest{
		ProfileID:      req.UserProfileID,
		ID:             req.ID,
		GroupExpenseID: req.GroupExpenseID,
		OtherFeeData: otherfee.OtherFeeData{
			Name:              req.Name,
			Amount:            req.Amount,
			CalculationMethod: req.CalculationMethod,
		},
	}

	otherFee, err := ofs.otherFeeClient.Update(ctx, request)
	if err != nil {
		return dto.OtherFeeResponse{}, err
	}

	profileIDs := []uuid.UUID{req.UserProfileID}
	profileIDs = append(profileIDs, otherFee.ProfileIDs()...)
	profilesByID, err := ofs.profileService.GetByIDs(ctx, profileIDs)
	if err != nil {
		return dto.OtherFeeResponse{}, err
	}

	return mapper.OtherFeeToResponse(otherFee, req.UserProfileID, profilesByID), nil
}

func (ofs *otherFeeServiceImpl) Remove(ctx context.Context, groupExpenseID, otherFeeID, userProfileID uuid.UUID) error {
	request := otherfee.RemoveRequest{
		ProfileID:      userProfileID,
		ID:             otherFeeID,
		GroupExpenseID: groupExpenseID,
	}

	return ofs.otherFeeClient.Remove(ctx, request)
}

func (ofs *otherFeeServiceImpl) GetCalculationMethods(ctx context.Context) ([]dto.FeeCalculationMethodInfo, error) {
	methods, err := ofs.otherFeeClient.GetCalculationMethods(ctx)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(methods, mapper.FeeCalculationMethodInfoToResponse), nil
}
