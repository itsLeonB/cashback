package fee

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ungerr"
)

type itemizedSplitFeeCalculator struct {
	method expenses.FeeCalculationMethod
}

func newItemizedSplitFeeCalculator() FeeCalculator {
	return &itemizedSplitFeeCalculator{
		expenses.ItemizedSplitFee,
	}
}

func (fc *itemizedSplitFeeCalculator) GetMethod() expenses.FeeCalculationMethod {
	return fc.method
}

func (fc *itemizedSplitFeeCalculator) Validate(fee expenses.OtherFee, groupExpense expenses.GroupExpense) error {
	if fee.ID == uuid.Nil {
		return ungerr.Unknown("fee ID cannot be nil")
	}

	if fee.GroupExpenseID == uuid.Nil {
		return ungerr.Unknown("group expense ID cannot be nil")
	}

	if fee.Amount.IsZero() {
		return ungerr.Unknown("amount cannot be zero")
	}

	if groupExpense.ItemsTotal.IsZero() {
		return ungerr.Unknown("items total cannot be zero")
	}

	if len(groupExpense.Participants) < 1 {
		return ungerr.Unknown("must have participants")
	}

	return nil
}

func (fc *itemizedSplitFeeCalculator) Split(fee expenses.OtherFee, groupExpense expenses.GroupExpense) []expenses.FeeParticipant {
	participantsCount := len(groupExpense.Participants)
	feeParticipants := make([]expenses.FeeParticipant, participantsCount)
	rate := fee.Amount.Div(groupExpense.ItemsTotal)

	for i, expenseParticipant := range groupExpense.Participants {
		feeParticipants[i] = expenses.FeeParticipant{
			OtherFeeID:  fee.ID,
			ProfileID:   expenseParticipant.ParticipantProfileID,
			ShareAmount: expenseParticipant.ShareAmount.Mul(rate),
		}
	}

	return feeParticipants
}

func (fc *itemizedSplitFeeCalculator) GetInfo() dto.FeeCalculationMethodInfo {
	return dto.FeeCalculationMethodInfo{
		Name:        fc.method,
		Display:     "Itemized split",
		Description: "Split the fee by a fixed rate applied to each expense items",
	}
}
