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
	billURL string,
	constructPreview bool,
) dto.GroupExpenseResponse {
	expense := dto.GroupExpenseResponse{
		BaseDTO:          BaseToDTO(groupExpense.BaseEntity),
		TotalAmount:      groupExpense.TotalAmount,
		ItemsTotalAmount: groupExpense.ItemsTotal,
		FeesTotalAmount:  groupExpense.FeesTotal,
		Description:      groupExpense.Description,
		Status:           groupExpense.Status,
		Payer:            ProfileToSimple(groupExpense.Payer, userProfileID),
		Creator:          ProfileToSimple(groupExpense.Creator, userProfileID),
		Items:            ezutil.MapSlice(groupExpense.Items, getExpenseItemSimpleMapper(userProfileID)),
		OtherFees:        ezutil.MapSlice(groupExpense.OtherFees, getOtherFeeSimpleMapper(userProfileID)),
		Participants:     ezutil.MapSlice(groupExpense.Participants, getExpenseParticipantSimpleMapper(userProfileID)),
		Bill:             ExpenseBillToResponse(groupExpense.Bill, billURL),
		BillExists:       groupExpense.Bill.ID != uuid.Nil,
	}

	if groupExpense.Description == "" {
		expense.Description = "Untitled Expense at " + time.Now().Format(time.DateOnly)
	}

	if constructPreview {
		expense.IsPreviewable = true
		expense.ConfirmationPreview = ToConfirmationResponse(groupExpense, userProfileID)
	}

	return expense
}

func GroupExpenseSimpleMapper(userProfileID uuid.UUID, billURL string, constructPreview bool) func(expenses.GroupExpense) dto.GroupExpenseResponse {
	return func(groupExpense expenses.GroupExpense) dto.GroupExpenseResponse {
		return GroupExpenseToResponse(groupExpense, userProfileID, billURL, constructPreview)
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
	debtTransactions := make([]debts.DebtTransaction, 0, len(groupExpense.Participants))
	for _, participant := range groupExpense.Participants {
		if groupExpense.PayerProfileID.UUID == participant.ParticipantProfileID {
			continue
		}
		debtTransactions = append(debtTransactions, debts.DebtTransaction{
			LenderProfileID:   groupExpense.PayerProfileID.UUID,
			BorrowerProfileID: participant.ParticipantProfileID,
			Amount:            participant.ShareAmount,
			TransferMethodID:  transferMethodID,
			Description:       fmt.Sprintf("Share for group expense %s: %s", groupExpense.ID, groupExpense.Description),
		})
	}

	return debtTransactions
}

func ToConfirmationResponse(expense expenses.GroupExpense, userProfileID uuid.UUID) dto.ExpenseConfirmationResponse {
	// Pre-compute participant maps for O(1) lookups
	itemParticipantMap := make(map[uuid.UUID]map[uuid.UUID]expenses.ItemParticipant)
	feeParticipantMap := make(map[uuid.UUID]map[uuid.UUID]expenses.FeeParticipant)

	// Build item participant lookup
	for _, item := range expense.Items {
		itemMap := make(map[uuid.UUID]expenses.ItemParticipant, len(item.Participants))
		for _, p := range item.Participants {
			itemMap[p.ProfileID] = p
		}
		itemParticipantMap[item.ID] = itemMap
	}

	// Build fee participant lookup
	for _, fee := range expense.OtherFees {
		feeMap := make(map[uuid.UUID]expenses.FeeParticipant, len(fee.Participants))
		for _, p := range fee.Participants {
			feeMap[p.ProfileID] = p
		}
		feeParticipantMap[fee.ID] = feeMap
	}

	participants := make([]dto.ConfirmedExpenseParticipant, len(expense.Participants))

	for i, participant := range expense.Participants {
		profileID := participant.ParticipantProfileID
		items := make([]dto.ConfirmedItemShare, 0, len(expense.Items))
		itemsTotal := decimal.Zero

		// Process items
		for _, item := range expense.Items {
			itemParticipant, ok := itemParticipantMap[item.ID][profileID]
			if !ok {
				continue
			}

			baseAmount := item.TotalAmount()
			shareAmount := itemParticipant.AllocatedAmount
			itemsTotal = itemsTotal.Add(shareAmount)

			// Calculate share rate for display purposes
			shareRate := decimal.Zero
			if baseAmount.IsPositive() {
				shareRate = shareAmount.Div(baseAmount)
			}

			items = append(items, dto.ConfirmedItemShare{
				ID:          item.ID,
				Name:        item.Name,
				BaseAmount:  baseAmount,
				ShareRate:   shareRate,
				ShareAmount: shareAmount,
			})
		}

		// Process fees
		fees := make([]dto.ConfirmedItemShare, 0, len(expense.OtherFees))
		feesTotal := decimal.Zero

		for _, fee := range expense.OtherFees {
			feeParticipant, ok := feeParticipantMap[fee.ID][profileID]
			if !ok {
				continue
			}

			feesTotal = feesTotal.Add(feeParticipant.ShareAmount)

			var shareRate decimal.Decimal
			if !feeParticipant.ShareAmount.IsZero() {
				shareRate = feeParticipant.ShareAmount.Div(fee.Amount)
			}

			fees = append(fees, dto.ConfirmedItemShare{
				ID:          fee.ID,
				Name:        fee.Name,
				BaseAmount:  fee.Amount,
				ShareRate:   shareRate,
				ShareAmount: feeParticipant.ShareAmount,
			})
		}

		participants[i] = dto.ConfirmedExpenseParticipant{
			Profile:    ProfileToSimple(participant.Profile, userProfileID),
			Items:      items,
			ItemsTotal: itemsTotal,
			Fees:       fees,
			FeesTotal:  feesTotal,
			Total:      participant.ShareAmount,
		}
	}

	return dto.ExpenseConfirmationResponse{
		ID:           expense.ID,
		Description:  expense.Description,
		TotalAmount:  expense.TotalAmount,
		Payer:        ProfileToSimple(expense.Payer, userProfileID),
		Participants: participants,
	}
}
