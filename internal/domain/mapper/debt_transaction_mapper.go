package mapper

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/shopspring/decimal"
	"golang.org/x/text/currency"
)

func MapToFriendBalanceSummary(userProfileID uuid.UUID, debtTransactions []dto.DebtTransactionResponse) dto.FriendBalance {
	totalOwedToYou, totalYouOwe := decimal.Zero, decimal.Zero

	for _, transaction := range debtTransactions {
		switch transaction.Type {
		case debts.Lend:
			switch transaction.Action {
			case debts.LendAction: // You lent money
				totalOwedToYou = totalOwedToYou.Add(transaction.Amount)
			case debts.BorrowAction: // You borrowed money
				totalYouOwe = totalYouOwe.Add(transaction.Amount)
			}
		case debts.Repay:
			switch transaction.Action {
			case debts.ReceiveAction: // You received repayment
				totalOwedToYou = totalOwedToYou.Sub(transaction.Amount)
			case debts.ReturnAction: // You returned money
				totalYouOwe = totalYouOwe.Sub(transaction.Amount)
			}
		}
	}

	return dto.FriendBalance{
		TotalOwedToYou: totalOwedToYou,
		TotalYouOwe:    totalYouOwe,
		NetBalance:     totalOwedToYou.Sub(totalYouOwe),
		CurrencyCode:   currency.IDR.String(),
	}
}
func DebtTransactionToResponse(userProfileID uuid.UUID, transaction debts.DebtTransaction) dto.DebtTransactionResponse {
	var profileID uuid.UUID
	if userProfileID == transaction.BorrowerProfileID && userProfileID != transaction.LenderProfileID {
		profileID = transaction.LenderProfileID
	} else if userProfileID == transaction.LenderProfileID && userProfileID != transaction.BorrowerProfileID {
		profileID = transaction.BorrowerProfileID
	}

	return dto.DebtTransactionResponse{
		BaseDTO:        BaseToDTO(transaction.BaseEntity),
		ProfileID:      profileID,
		Type:           transaction.Type,
		Action:         transaction.Action,
		Amount:         transaction.Amount,
		TransferMethod: transaction.TransferMethod.Display,
		Description:    transaction.Description,
	}
}

func DebtTransactionSimpleMapper(userProfileID uuid.UUID) func(debts.DebtTransaction) dto.DebtTransactionResponse {
	return func(transaction debts.DebtTransaction) dto.DebtTransactionResponse {
		return DebtTransactionToResponse(userProfileID, transaction)
	}
}
