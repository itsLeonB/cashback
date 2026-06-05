package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/shopspring/decimal"
)

func MapToFriendBalanceSummary(transactions []debts.DebtTransaction, userAssociatedIDs []uuid.UUID) dto.FriendBalance {
	totalLent, totalBorrowed, history := calculateBalances(userAssociatedIDs, transactions)

	return dto.FriendBalance{
		NetBalance:              totalLent.Sub(totalBorrowed),
		TotalLentToFriend:       totalLent,
		TotalBorrowedFromFriend: totalBorrowed,
		TransactionHistory:      history,
	}
}

func SummarizePerCurrency(transactions []debts.DebtTransaction, userAssociatedIDs []uuid.UUID) map[string]dto.FriendBalance {
	transactionsByCurrency := make(map[string][]debts.DebtTransaction)
	for _, transaction := range transactions {
		transactionsByCurrency[transaction.Currency] = append(transactionsByCurrency[transaction.Currency], transaction)
	}

	balancesPerCurrency := make(map[string]dto.FriendBalance, len(transactionsByCurrency))
	for currency, transactions := range transactionsByCurrency {
		balancesPerCurrency[currency] = MapToFriendBalanceSummary(transactions, userAssociatedIDs)
	}

	return balancesPerCurrency
}

func calculateBalances(userAssociatedIDs []uuid.UUID, transactions []debts.DebtTransaction) (decimal.Decimal, decimal.Decimal, []dto.FriendTransactionItem) {
	totalLent, totalBorrowed := decimal.Zero, decimal.Zero
	history := make([]dto.FriendTransactionItem, 0, len(transactions))

	userIDMap := buildIDSet(userAssociatedIDs)

	for _, tx := range transactions {
		dir := classifyTransaction(tx, userIDMap)
		if dir == 0 {
			logger.Errorf("orphaned transaction %s", tx.ID)
			continue
		}

		var transactionType string
		var amount decimal.Decimal
		if dir > 0 {
			transactionType = "LENT"
			amount = tx.Amount
			totalLent = totalLent.Add(tx.Amount)
		} else {
			transactionType = "BORROWED"
			amount = tx.Amount
			totalBorrowed = totalBorrowed.Add(tx.Amount)
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

// buildIDSet creates a set for fast lookup.
func buildIDSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	s := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

// classifyTransaction returns +1 if user is lender, -1 if borrower, 0 if ambiguous.
func classifyTransaction(tx debts.DebtTransaction, userIDSet map[uuid.UUID]struct{}) int {
	_, userIsLender := userIDSet[tx.LenderProfileID]
	_, userIsBorrower := userIDSet[tx.BorrowerProfileID]
	switch {
	case userIsLender && !userIsBorrower:
		return 1
	case userIsBorrower && !userIsLender:
		return -1
	default:
		return 0
	}
}

func DebtTransactionToResponse(userProfileID uuid.UUID, transaction debts.DebtTransaction, profilesByID map[uuid.UUID]dto.ProfileResponse) dto.DebtTransactionResponse {
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
		BaseDTO: BaseToDTO(transaction.BaseEntity),
		Profile: dto.SimpleProfile{
			ID:     profileID,
			Name:   profilesByID[profileID].Name,
			Avatar: profilesByID[profileID].Avatar,
			IsUser: profileID == userProfileID,
		},
		Type:           txType,
		Currency:       transaction.Currency,
		Amount:         transaction.Amount,
		TransferMethod: transaction.TransferMethod.Display,
		Description:    transaction.Description,
		GroupExpenseID: transaction.GroupExpenseID.UUID,
		IsFromExpense:  transaction.GroupExpenseID.Valid,
	}
}

func DebtTransactionSimpleMapper(userProfileID uuid.UUID, profilesByID map[uuid.UUID]dto.ProfileResponse) func(debts.DebtTransaction) dto.DebtTransactionResponse {
	return func(transaction debts.DebtTransaction) dto.DebtTransactionResponse {
		return DebtTransactionToResponse(userProfileID, transaction, profilesByID)
	}
}

// NetBalanceByFriend groups transactions by counterparty and computes net balance per currency.
// Returns map[counterpartyProfileID]map[currency]netBalance.
func NetBalanceByFriend(transactions []debts.DebtTransaction, userAssociatedIDs []uuid.UUID) map[uuid.UUID]map[string]decimal.Decimal {
	userIDSet := buildIDSet(userAssociatedIDs)

	result := make(map[uuid.UUID]map[string]decimal.Decimal)
	for _, tx := range transactions {
		dir := classifyTransaction(tx, userIDSet)
		if dir == 0 {
			continue
		}

		var counterparty uuid.UUID
		var amount decimal.Decimal
		if dir > 0 {
			counterparty = tx.BorrowerProfileID
			amount = tx.Amount
		} else {
			counterparty = tx.LenderProfileID
			amount = tx.Amount.Neg()
		}

		if result[counterparty] == nil {
			result[counterparty] = make(map[string]decimal.Decimal)
		}
		result[counterparty][tx.Currency] = result[counterparty][tx.Currency].Add(amount)
	}

	return result
}
