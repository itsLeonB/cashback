package mapper

import (
	"fmt"

	"github.com/google/uuid"
	expense "github.com/itsLeonB/billsplittr-protos/gen/go/groupexpense/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/domain/debt"
	"github.com/itsLeonB/orcashtrator/internal/domain/expenseitem"
	"github.com/itsLeonB/orcashtrator/internal/domain/groupexpense"
	"github.com/itsLeonB/orcashtrator/internal/domain/otherfee"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/ungerr"
	"github.com/rotisserie/eris"
	"github.com/shopspring/decimal"
)

func GroupExpenseRequestToEntity(request dto.NewGroupExpenseRequest) groupexpense.CreateDraftRequest {
	return groupexpense.CreateDraftRequest{
		PayerProfileID: request.PayerProfileID,
		TotalAmount:    request.TotalAmount,
		Subtotal:       request.Subtotal,
		Description:    request.Description,
		Items:          ezutil.MapSlice(request.Items, NewExpenseItemRequestToData),
		OtherFees:      ezutil.MapSlice(request.OtherFees, otherFeeRequestToData),
	}
}

func GroupExpenseToResponse(
	groupExpense groupexpense.GroupExpense,
	userProfileID uuid.UUID,
	profilesByID map[uuid.UUID]dto.ProfileResponse,
	billResponse dto.ExpenseBillResponse,
) dto.GroupExpenseResponse {
	return dto.GroupExpenseResponse{
		ID:                    groupExpense.ID,
		TotalAmount:           groupExpense.TotalAmount,
		ItemsTotalAmount:      groupExpense.ItemsTotal,
		FeesTotalAmount:       groupExpense.FeesTotal,
		Description:           groupExpense.Description,
		Confirmed:             groupExpense.IsConfirmed,
		ParticipantsConfirmed: groupExpense.IsParticipantsConfirmed,
		Status:                groupExpense.Status,
		CreatedAt:             groupExpense.CreatedAt,
		UpdatedAt:             groupExpense.UpdatedAt,
		DeletedAt:             groupExpense.DeletedAt,
		Payer:                 SimpleProfileMapper(userProfileID)(profilesByID[groupExpense.PayerProfileID]),
		Creator:               SimpleProfileMapper(userProfileID)(profilesByID[groupExpense.CreatorProfileID]),
		Items:                 ezutil.MapSlice(groupExpense.Items, getExpenseItemSimpleMapper(userProfileID, profilesByID)),
		OtherFees:             ezutil.MapSlice(groupExpense.OtherFees, getOtherFeeSimpleMapper(userProfileID, profilesByID)),
		Participants:          ezutil.MapSlice(groupExpense.Participants, getExpenseParticipantSimpleMapper(userProfileID, profilesByID)),
		Bill:                  billResponse,
		BillExists:            billResponse.ID != uuid.Nil,
	}
}

func getExpenseItemSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(item expenseitem.ExpenseItem) dto.ExpenseItemResponse {
	return func(item expenseitem.ExpenseItem) dto.ExpenseItemResponse {
		return ExpenseItemToResponse(item, userProfileID, profilesByID)
	}
}

func ExpenseItemToResponse(item expenseitem.ExpenseItem, userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) dto.ExpenseItemResponse {
	return dto.ExpenseItemResponse{
		ID:             item.ID,
		GroupExpenseID: item.GroupExpenseID,
		Name:           item.Name,
		Amount:         item.Amount,
		Quantity:       item.Quantity,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
		DeletedAt:      item.DeletedAt,
		Participants:   ezutil.MapSlice(item.Participants, getItemParticipantSimpleMapper(userProfileID, profilesByID)),
	}
}

func getOtherFeeSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(otherfee.OtherFee) dto.OtherFeeResponse {
	return func(fee otherfee.OtherFee) dto.OtherFeeResponse {
		return OtherFeeToResponse(fee, userProfileID, profilesByID)
	}
}

func OtherFeeToResponse(fee otherfee.OtherFee, userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) dto.OtherFeeResponse {
	return dto.OtherFeeResponse{
		ID:                fee.ID,
		Name:              fee.Name,
		Amount:            fee.Amount,
		CalculationMethod: fee.CalculationMethod,
		CreatedAt:         fee.CreatedAt,
		UpdatedAt:         fee.UpdatedAt,
		DeletedAt:         fee.DeletedAt,
		Participants:      ezutil.MapSlice(fee.Participants, getFeeParticipantSimpleMapper(userProfileID, profilesByID)),
	}
}

func getFeeParticipantSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(otherfee.FeeParticipant) dto.FeeParticipantResponse {
	return func(feeParticipant otherfee.FeeParticipant) dto.FeeParticipantResponse {
		return feeParticipantToResponse(feeParticipant, userProfileID, profilesByID[feeParticipant.ProfileID])
	}
}

func feeParticipantToResponse(feeParticipant otherfee.FeeParticipant, userProfileID uuid.UUID, profile dto.ProfileResponse) dto.FeeParticipantResponse {
	return dto.FeeParticipantResponse{
		Profile:     ToSimpleProfile(profile, userProfileID),
		ShareAmount: feeParticipant.ShareAmount,
	}
}

func getItemParticipantSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(itemParticipant expenseitem.ItemParticipant) dto.ItemParticipantResponse {
	return func(itemParticipant expenseitem.ItemParticipant) dto.ItemParticipantResponse {
		return itemParticipantToResponse(itemParticipant, userProfileID, profilesByID[itemParticipant.ProfileID])
	}
}

func itemParticipantToResponse(itemParticipant expenseitem.ItemParticipant, userProfileID uuid.UUID, profile dto.ProfileResponse) dto.ItemParticipantResponse {
	return dto.ItemParticipantResponse{
		Profile:    ToSimpleProfile(profile, userProfileID),
		ShareRatio: itemParticipant.Share,
	}
}

func otherFeeRequestToData(req dto.NewOtherFeeRequest) otherfee.OtherFeeData {
	return otherfee.OtherFeeData{
		Name:              req.Name,
		Amount:            req.Amount,
		CalculationMethod: req.CalculationMethod,
	}
}

func ExpenseParticipantToResponse(expenseParticipant groupexpense.ExpenseParticipant, userProfileID uuid.UUID, profile dto.ProfileResponse) dto.ExpenseParticipantResponse {
	return dto.ExpenseParticipantResponse{
		Profile:     ToSimpleProfile(profile, userProfileID),
		ShareAmount: expenseParticipant.ShareAmount,
	}
}

func getExpenseParticipantSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(groupexpense.ExpenseParticipant) dto.ExpenseParticipantResponse {
	return func(ep groupexpense.ExpenseParticipant) dto.ExpenseParticipantResponse {
		return ExpenseParticipantToResponse(ep, userProfileID, profilesByID[ep.ProfileID])
	}
}

func ExpenseParticipantToData(participant groupexpense.ExpenseParticipant) (debt.ExpenseParticipantData, error) {
	if participant.ShareAmount.LessThanOrEqual(decimal.Zero) {
		return debt.ExpenseParticipantData{}, ungerr.UnprocessableEntityError(fmt.Sprintf(
			"participant %s has share amount: %s",
			participant.ProfileID,
			participant.ShareAmount.String(),
		))
	}
	return debt.ExpenseParticipantData{
		ProfileID:   participant.ProfileID,
		ShareAmount: participant.ShareAmount,
	}, nil
}

func FromExpenseStatusProto(status expense.GroupExpenseResponse_Status) (appconstant.ExpenseStatus, error) {
	switch status {
	case expense.GroupExpenseResponse_STATUS_UNSPECIFIED:
		return "", eris.New("unspecified expense status")
	case expense.GroupExpenseResponse_STATUS_DRAFT:
		return appconstant.DraftExpense, nil
	case expense.GroupExpenseResponse_STATUS_READY:
		return appconstant.ReadyExpense, nil
	case expense.GroupExpenseResponse_STATUS_CONFIRMED:
		return appconstant.ConfirmedExpense, nil
	default:
		return "", eris.Errorf("unknown expense status enum: %s", status.String())
	}
}
