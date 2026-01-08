package debt

import (
	"fmt"

	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
)

type DebtCalculator interface {
	GetAction() debts.DebtTransactionAction
	MapRequestToEntity(request dto.NewDebtTransactionRequest) debts.DebtTransaction
	MapEntityToResponse(debtTransaction debts.DebtTransaction) dto.DebtTransactionResponse
}

var initFuncs = []func() DebtCalculator{
	newBorrowingDebtCalculator,
	newLendingDebtCalculator,
	newReceivingDebtCalculator,
	newReturningDebtCalculator,
}

func NewDebtCalculatorStrategies() map[debts.DebtTransactionAction]DebtCalculator {
	strategyMap := make(map[debts.DebtTransactionAction]DebtCalculator)

	for _, initFunc := range initFuncs {
		if initFunc == nil {
			panic("initFunc is nil")
		}

		calculator := initFunc()
		if calculator == nil {
			panic("calculator is nil")
		}

		action := calculator.GetAction()
		if _, exists := strategyMap[action]; exists {
			panic(fmt.Sprintf("duplicate calculator for action: %s", action))
		}

		strategyMap[calculator.GetAction()] = calculator
	}

	return strategyMap
}
