package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

func MapToFriendBalanceSummary(transactions []debts.DebtTransaction, userAssociatedIDs []uuid.UUID) dto.FriendBalance {
	totalLent, totalBorrowed, history := calculateBalances(userAssociatedIDs, transactions)

	return dto.FriendBalance{
		NetBalance:              totalLent.Sub(totalBorrowed),
		TotalLentToFriend:       totalLent,
		TotalBorrowedFromFriend: totalBorrowed,
		TransactionHistory:      history,
		CurrencyCode:            currency.IDR.String(),
	}
}

func calculateBalances(userAssociatedIDs []uuid.UUID, transactions []debts.DebtTransaction) (decimal.Decimal, decimal.Decimal, []dto.FriendTransactionItem) {
	totalLent, totalBorrowed := decimal.Zero, decimal.Zero
	history := make([]dto.FriendTransactionItem, 0, len(transactions))

	// Create a map for quick lookup of user's associated IDs
	userIDMap := make(map[uuid.UUID]struct{}, len(userAssociatedIDs))
	for _, id := range userAssociatedIDs {
		userIDMap[id] = struct{}{}
	}

	for _, tx := range transactions {
		var transactionType string
		var amount decimal.Decimal

		// Check if user (or their associated profiles) is the lender or borrower
		_, userIsLender := userIDMap[tx.LenderProfileID]
		_, userIsBorrower := userIDMap[tx.BorrowerProfileID]

		if userIsLender && !userIsBorrower {
			// User is the lender
			transactionType = "LENT"
			amount = tx.Amount
			totalLent = totalLent.Add(tx.Amount)
		} else if userIsBorrower && !userIsLender {
			// User is the borrower
			transactionType = "BORROWED"
			amount = tx.Amount
			totalBorrowed = totalBorrowed.Add(tx.Amount)
		} else {
			// Skip transactions where user is both or neither (shouldn't happen)
			logger.Errorf("orphaned transaction %s. userIsLender: %t. userIsBorrower: %t", tx.ID, userIsLender, userIsBorrower)
			continue
		}

		history = append(history, dto.FriendTransactionItem{
			BaseDTO:        BaseToDTO(tx.BaseEntity),
			Type:           transactionType,
			Amount:         amount,
			TransferMethod: tx.TransferMethod.Display,
			Description:    tx.Description,
		})
	}

	return totalLent, totalBorrowed, history
}

func DebtTransactionToResponse(userProfileID uuid.UUID, transaction debts.DebtTransaction) dto.DebtTransactionResponse {
	var profileID uuid.UUID
	var txType string

	if userProfileID == transaction.BorrowerProfileID {
		profileID = transaction.LenderProfileID
		txType = "BORROWED"
	} else {
		profileID = transaction.BorrowerProfileID
		txType = "LENT"
	}

	return dto.DebtTransactionResponse{
		BaseDTO:        BaseToDTO(transaction.BaseEntity),
		ProfileID:      profileID,
		Type:           txType,
		Amount:         transaction.Amount,
		TransferMethod: transaction.TransferMethod.Display,
		Description:    transaction.Description,
		GroupExpenseID: transaction.GroupExpenseID.UUID,
		IsFromExpense:  transaction.GroupExpenseID.Valid,
	}
}

func DebtTransactionSimpleMapper(userProfileID uuid.UUID) func(debts.DebtTransaction) dto.DebtTransactionResponse {
	return func(transaction debts.DebtTransaction) dto.DebtTransactionResponse {
		return DebtTransactionToResponse(userProfileID, transaction)
	}
}
