package subscriber

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/meq/task"
	"github.com/itsLeonB/ungerr"
)

func expenseBillTextExtractedHandler(expenseSvc service.GroupExpenseService) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		var taskMsg task.Task[message.ExpenseBillTextExtracted]
		if err := json.Unmarshal(t.Payload(), &taskMsg); err != nil {
			return ungerr.Wrapf(err, "error unmarshaling payload to: %T", taskMsg)
		}

		return expenseSvc.ParseFromBillText(ctx, taskMsg.Message)
	})
}
