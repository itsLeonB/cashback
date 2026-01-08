package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ezutil/v2"
)

func OtherFeeRequestToEntity(request dto.NewOtherFeeRequest) expenses.OtherFee {
	return expenses.OtherFee{
		GroupExpenseID:    request.GroupExpenseID,
		Name:              request.Name,
		Amount:            request.Amount,
		CalculationMethod: request.CalculationMethod,
	}
}

func PatchOtherFeeWithRequest(otherFee expenses.OtherFee, request dto.UpdateOtherFeeRequest) expenses.OtherFee {
	otherFee.Name = request.Name
	otherFee.Amount = request.Amount
	otherFee.CalculationMethod = request.CalculationMethod
	return otherFee
}

func getOtherFeeSimpleMapper(userProfileID uuid.UUID) func(expenses.OtherFee) dto.OtherFeeResponse {
	return func(fee expenses.OtherFee) dto.OtherFeeResponse {
		return OtherFeeToResponse(fee, userProfileID)
	}
}

func OtherFeeToResponse(fee expenses.OtherFee, userProfileID uuid.UUID) dto.OtherFeeResponse {
	return dto.OtherFeeResponse{
		BaseDTO:           BaseToDTO(fee.BaseEntity),
		Name:              fee.Name,
		Amount:            fee.Amount,
		CalculationMethod: fee.CalculationMethod,
		Participants:      ezutil.MapSlice(fee.Participants, getFeeParticipantSimpleMapper(userProfileID)),
	}
}

func getFeeParticipantSimpleMapper(userProfileID uuid.UUID) func(expenses.FeeParticipant) dto.FeeParticipantResponse {
	return func(feeParticipant expenses.FeeParticipant) dto.FeeParticipantResponse {
		return feeParticipantToResponse(feeParticipant, userProfileID)
	}
}

func feeParticipantToResponse(feeParticipant expenses.FeeParticipant, userProfileID uuid.UUID) dto.FeeParticipantResponse {
	return dto.FeeParticipantResponse{
		Profile:     ProfileToSimple(feeParticipant.Profile, userProfileID),
		ShareAmount: feeParticipant.ShareAmount,
	}
}

func otherFeeRequestToData(req dto.NewOtherFeeRequest) expenses.OtherFee {
	return expenses.OtherFee{
		Name:              req.Name,
		Amount:            req.Amount,
		CalculationMethod: req.CalculationMethod,
	}
}
