package mapper

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

func GroupExpenseRequestToEntity(request dto.NewGroupExpenseRequest) expenses.GroupExpense {
	expense := expenses.GroupExpense{
		TotalAmount: request.TotalAmount,
		ItemsTotal:  request.Subtotal,
		FeesTotal:   request.TotalAmount.Sub(request.Subtotal),
		Description: request.Description,
		Items:       ezutil.MapSlice(request.Items, NewExpenseItemRequestToData),
		OtherFees:   ezutil.MapSlice(request.OtherFees, otherFeeRequestToData),
	}

	if request.PayerProfileID != uuid.Nil {
		expense.PayerProfileID = uuid.NullUUID{
			UUID:  request.PayerProfileID,
			Valid: true,
		}
	}

	return expense
}

func GroupExpenseToResponse(
	groupExpense expenses.GroupExpense,
	userProfileID uuid.UUID,
) dto.GroupExpenseResponse {
	description := groupExpense.Description
	if description == "" {
		description = "Untitled Expense at " + time.Now().Format(time.DateOnly)
	}
	return dto.GroupExpenseResponse{
		BaseDTO:          BaseToDTO(groupExpense.BaseEntity),
		TotalAmount:      groupExpense.TotalAmount,
		ItemsTotalAmount: groupExpense.ItemsTotal,
		FeesTotalAmount:  groupExpense.FeesTotal,
		Description:      description,
		Status:           groupExpense.Status,
		Payer:            ProfileToSimple(groupExpense.Payer, userProfileID),
		Creator:          ProfileToSimple(groupExpense.Creator, userProfileID),
		Items:            ezutil.MapSlice(groupExpense.Items, getExpenseItemSimpleMapper(userProfileID)),
		OtherFees:        ezutil.MapSlice(groupExpense.OtherFees, getOtherFeeSimpleMapper(userProfileID)),
		Participants:     ezutil.MapSlice(groupExpense.Participants, getExpenseParticipantSimpleMapper(userProfileID)),
		Bill:             ExpenseBillToResponse(groupExpense.Bill),
		BillExists:       groupExpense.Bill.ID != uuid.Nil,
	}
}

func GroupExpenseSimpleMapper(userProfileID uuid.UUID) func(expenses.GroupExpense) dto.GroupExpenseResponse {
	return func(groupExpense expenses.GroupExpense) dto.GroupExpenseResponse {
		return GroupExpenseToResponse(groupExpense, userProfileID)
	}
}

func getExpenseItemSimpleMapper(userProfileID uuid.UUID) func(item expenses.ExpenseItem) dto.ExpenseItemResponse {
	return func(item expenses.ExpenseItem) dto.ExpenseItemResponse {
		return ExpenseItemToResponse(item, userProfileID)
	}
}

func ExpenseItemToResponse(item expenses.ExpenseItem, userProfileID uuid.UUID) dto.ExpenseItemResponse {
	return dto.ExpenseItemResponse{
		BaseDTO:        BaseToDTO(item.BaseEntity),
		GroupExpenseID: item.GroupExpenseID,
		Name:           item.Name,
		Amount:         item.Amount,
		Quantity:       item.Quantity,
		Participants:   ezutil.MapSlice(item.Participants, getItemParticipantSimpleMapper(userProfileID)),
	}
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

func getItemParticipantSimpleMapper(userProfileID uuid.UUID) func(itemParticipant expenses.ItemParticipant) dto.ItemParticipantResponse {
	return func(itemParticipant expenses.ItemParticipant) dto.ItemParticipantResponse {
		return itemParticipantToResponse(itemParticipant, userProfileID)
	}
}

func itemParticipantToResponse(itemParticipant expenses.ItemParticipant, userProfileID uuid.UUID) dto.ItemParticipantResponse {
	return dto.ItemParticipantResponse{
		Profile:    ProfileToSimple(itemParticipant.Profile, userProfileID),
		ShareRatio: itemParticipant.Share,
	}
}

func otherFeeRequestToData(req dto.NewOtherFeeRequest) expenses.OtherFee {
	return expenses.OtherFee{
		Name:              req.Name,
		Amount:            req.Amount,
		CalculationMethod: req.CalculationMethod,
	}
}

func ExpenseParticipantToResponse(expenseParticipant expenses.ExpenseParticipant, userProfileID uuid.UUID) dto.ExpenseParticipantResponse {
	return dto.ExpenseParticipantResponse{
		Profile:     ProfileToSimple(expenseParticipant.Profile, userProfileID),
		ShareAmount: expenseParticipant.ShareAmount,
	}
}

func getExpenseParticipantSimpleMapper(userProfileID uuid.UUID) func(expenses.ExpenseParticipant) dto.ExpenseParticipantResponse {
	return func(ep expenses.ExpenseParticipant) dto.ExpenseParticipantResponse {
		return ExpenseParticipantToResponse(ep, userProfileID)
	}
}

func ExpenseParticipantToData(participant expenses.ExpenseParticipant) (expenses.ExpenseParticipant, error) {
	if participant.ShareAmount.LessThanOrEqual(decimal.Zero) {
		return expenses.ExpenseParticipant{}, ungerr.UnprocessableEntityError(fmt.Sprintf(
			"participant %s has share amount: %s",
			participant.ParticipantProfileID,
			participant.ShareAmount.String(),
		))
	}
	return expenses.ExpenseParticipant{
		ParticipantProfileID: participant.ParticipantProfileID,
		ShareAmount:          participant.ShareAmount,
	}, nil
}

func GroupExpenseToDebtTransactions(groupExpense expenses.GroupExpense, transferMethodID uuid.UUID) []debts.DebtTransaction {
	action := debts.BorrowAction
	if groupExpense.PayerProfileID.UUID == groupExpense.CreatorProfileID {
		action = debts.LendAction
	}

	debtTransactions := make([]debts.DebtTransaction, 0, len(groupExpense.Participants))
	for _, participant := range groupExpense.Participants {
		if groupExpense.PayerProfileID.UUID == participant.ParticipantProfileID {
			continue
		}
		debtTransactions = append(debtTransactions, debts.DebtTransaction{
			LenderProfileID:   groupExpense.PayerProfileID.UUID,
			BorrowerProfileID: participant.ParticipantProfileID,
			Type:              debts.Lend,
			Action:            action,
			Amount:            participant.ShareAmount,
			TransferMethodID:  transferMethodID,
			Description:       fmt.Sprintf("Share for group expense %s: %s", groupExpense.ID, groupExpense.Description),
		})
	}

	return debtTransactions
}
