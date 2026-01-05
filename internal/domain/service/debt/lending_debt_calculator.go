package debt

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

type lendingDebtCalculator struct {
	action debts.DebtTransactionAction
}

func newLendingDebtCalculator() DebtCalculator {
	return &lendingDebtCalculator{
		action: debts.LendAction,
	}
}

func (dc *lendingDebtCalculator) GetAction() debts.DebtTransactionAction {
	return dc.action
}

func (dc *lendingDebtCalculator) MapRequestToEntity(request dto.NewDebtTransactionRequest) debts.DebtTransaction {
	return debts.DebtTransaction{
		LenderProfileID:   request.UserProfileID,
		BorrowerProfileID: request.FriendProfileID,
		Type:              debts.Lend,
		Action:            dc.action,
		Amount:            request.Amount,
		TransferMethodID:  request.TransferMethodID,
		Description:       request.Description,
	}
}

func (dc *lendingDebtCalculator) MapEntityToResponse(debtTransaction debts.DebtTransaction) dto.DebtTransactionResponse {
	return dto.DebtTransactionResponse{
		BaseDTO:        mapper.BaseToDTO(debtTransaction.BaseEntity),
		ProfileID:      debtTransaction.BorrowerProfileID,
		Type:           debtTransaction.Type,
		Action:         debtTransaction.Action,
		Amount:         debtTransaction.Amount,
		TransferMethod: debtTransaction.TransferMethod.Display,
		Description:    debtTransaction.Description,
	}
}
