package expense

import (
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/ungerr"
	"github.com/shopspring/decimal"
)

type CalculationService interface {
	RecalculateExpense(expense expenses.GroupExpense, amountChanged bool) (expenses.GroupExpense, bool, error)
}

func NewCalculationService() *calculationServiceImpl {
	return &calculationServiceImpl{}
}

type calculationServiceImpl struct{}

// This is a stateless service to efficiently and effectively recalculate a GroupExpense's ItemsTotal and Total Amount.
// Intended to be used when introducing changes to ExpenseItem list, whether adding/removing, or updating amount/participants.
// To make the calculation efficient, amountChanged value is used as a hint, whether to do a recalculation or not.
func (c *calculationServiceImpl) RecalculateExpense(expense expenses.GroupExpense, amountChanged bool) (expenses.GroupExpense, bool, error) {
	itemsTotal := decimal.Zero
	allItemsHaveParticipants := len(expense.Items) > 0

	// Single pass: calculate totals and check participant status
	for _, item := range expense.Items {
		if amountChanged {
			itemsTotal = itemsTotal.Add(item.TotalAmount())
		}
		if len(item.Participants) == 0 {
			allItemsHaveParticipants = false
		}
	}

	// Determine new status
	newStatus := expenses.DraftExpense
	if len(expense.Items) > 0 && allItemsHaveParticipants {
		newStatus = expenses.ReadyExpense
	}

	// Check if we need to update the group expense
	statusChanged := expense.Status != newStatus
	if amountChanged || statusChanged {
		// Update group expense fields
		expense.Status = newStatus

		if amountChanged {
			expense.ItemsTotal = itemsTotal
			expense.TotalAmount = itemsTotal.Add(expense.FeesTotal)

			if expense.TotalAmount.IsNegative() {
				return expenses.GroupExpense{}, false, ungerr.UnprocessableEntityError("total group expense amount must be positive")
			}
		}
	}

	return expense, amountChanged || statusChanged, nil
}
