package fee

import (
	"log"

	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
)

var namespace = "[FeeCalculator]"

type FeeCalculator interface {
	GetMethod() expenses.FeeCalculationMethod
	Validate(fee expenses.OtherFee, groupExpense expenses.GroupExpense) error
	Split(fee expenses.OtherFee, groupExpense expenses.GroupExpense) []expenses.FeeParticipant
	GetInfo() dto.FeeCalculationMethodInfo
}

var initFuncs = []func() FeeCalculator{
	newEqualSplitFeeCalculator,
	newItemizedSplitFeeCalculator,
}

func NewFeeCalculatorRegistry() map[expenses.FeeCalculationMethod]FeeCalculator {
	registry := make(map[expenses.FeeCalculationMethod]FeeCalculator)

	for _, initFunc := range initFuncs {
		if initFunc == nil {
			log.Fatalf("%s initFunc is nil", namespace)
		}

		calculator := initFunc()
		if calculator == nil {
			log.Fatalf("%s calculator is nil", namespace)
		}

		method := calculator.GetMethod()
		if _, exists := registry[method]; exists {
			log.Fatalf("%s duplicate calculator for method: %s", namespace, method)
		}

		registry[calculator.GetMethod()] = calculator
	}

	return registry
}
