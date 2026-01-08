package debt

import (
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
)

type returningDebtCalculator struct {
	action debts.DebtTransactionAction
}

func newReturningDebtCalculator() DebtCalculator {
	return &returningDebtCalculator{
		action: debts.ReturnAction,
	}
}

func (dc *returningDebtCalculator) GetAction() debts.DebtTransactionAction {
	return dc.action
}

func (dc *returningDebtCalculator) MapRequestToEntity(request dto.NewDebtTransactionRequest) debts.DebtTransaction {
	return debts.DebtTransaction{
		LenderProfileID:   request.FriendProfileID,
		BorrowerProfileID: request.UserProfileID,
		Action:            dc.action,
		Type:              debts.Repay,
		Amount:            request.Amount,
		TransferMethodID:  request.TransferMethodID,
		Description:       request.Description,
	}
}

func (dc *returningDebtCalculator) MapEntityToResponse(debtTransaction debts.DebtTransaction) dto.DebtTransactionResponse {
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
