package subscriber

import (
	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
)

func expenseBillTextExtractedHandler(expenseSvc service.GroupExpenseService) asynq.Handler {
	return withLogging(message.ExpenseBillTextExtracted{}.Type(), expenseSvc.ParseFromBillText)
}
