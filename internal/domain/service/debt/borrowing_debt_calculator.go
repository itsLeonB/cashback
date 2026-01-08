package debt

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

type borrowingDebtCalculator struct {
	action debts.DebtTransactionAction
}

func newBorrowingDebtCalculator() DebtCalculator {
	return &borrowingDebtCalculator{
		action: debts.BorrowAction,
	}
}

func (dc *borrowingDebtCalculator) GetAction() debts.DebtTransactionAction {
	return dc.action
}

func (dc *borrowingDebtCalculator) MapRequestToEntity(request dto.NewDebtTransactionRequest) debts.DebtTransaction {
	return debts.DebtTransaction{
		LenderProfileID:   request.FriendProfileID,
		BorrowerProfileID: request.UserProfileID,
		Type:              debts.Lend,
		Action:            dc.action,
		Amount:            request.Amount,
		TransferMethodID:  request.TransferMethodID,
		Description:       request.Description,
	}
}

func (dc *borrowingDebtCalculator) MapEntityToResponse(debtTransaction debts.DebtTransaction) dto.DebtTransactionResponse {
	return dto.DebtTransactionResponse{
		BaseDTO:        mapper.BaseToDTO(debtTransaction.BaseEntity),
		ProfileID:      debtTransaction.LenderProfileID,
		Type:           debtTransaction.Type,
		Action:         debtTransaction.Action,
		Amount:         debtTransaction.Amount,
		TransferMethod: debtTransaction.TransferMethod.Display,
		Description:    debtTransaction.Description,
	}
}
