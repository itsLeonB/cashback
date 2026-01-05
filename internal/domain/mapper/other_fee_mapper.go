package mapper

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
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
