package fee

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type equalSplitFeeCalculator struct {
	method expenses.FeeCalculationMethod
}

func newEqualSplitFeeCalculator() FeeCalculator {
	return &equalSplitFeeCalculator{
		expenses.EqualSplitFee,
	}
}

func (fc *equalSplitFeeCalculator) GetMethod() expenses.FeeCalculationMethod {
	return fc.method
}

func (fc *equalSplitFeeCalculator) Validate(fee expenses.OtherFee, groupExpense expenses.GroupExpense) error {
	if fee.ID == uuid.Nil {
		return ungerr.Unknown("fee ID cannot be nil")
	}

	if fee.GroupExpenseID == uuid.Nil {
		return ungerr.Unknown("group expense ID cannot be nil")
	}

	if fee.Amount.IsZero() {
		return ungerr.Unknown("amount cannot be zero")
	}

	if len(groupExpense.Participants) < 1 {
		return ungerr.Unknown("must have participants")
	}

	return nil
}

func (fc *equalSplitFeeCalculator) Split(fee expenses.OtherFee, groupExpense expenses.GroupExpense) []expenses.FeeParticipant {
	participantsCount := len(groupExpense.Participants)
	feeParticipants := make([]expenses.FeeParticipant, participantsCount)
	amountPerParticipant := fee.Amount.Div(decimal.NewFromInt(int64(participantsCount)))

	for i, expenseParticipant := range groupExpense.Participants {
		feeParticipants[i] = expenses.FeeParticipant{
			OtherFeeID:  fee.ID,
			ProfileID:   expenseParticipant.ParticipantProfileID,
			ShareAmount: amountPerParticipant,
		}
	}

	return feeParticipants
}

func (fc *equalSplitFeeCalculator) GetInfo() dto.FeeCalculationMethodInfo {
	return dto.FeeCalculationMethodInfo{
		Name:        fc.method,
		Display:     "Equal split",
		Description: "Equally split the fee to all participants",
	}
}
