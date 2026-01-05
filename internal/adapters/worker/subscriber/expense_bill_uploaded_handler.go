package subscriber

import (
	"context"
	"encoding/json"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/meq/task"
	"github.com/itsLeonB/ungerr"
)

func expenseBillUploadedHandler(
	billSvc service.ExpenseBillService,
	extractedQueue meq.TaskQueue[message.ExpenseBillTextExtracted],
) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		var taskMsg task.Task[message.ExpenseBillUploaded]
		if err := json.Unmarshal(t.Payload(), &taskMsg); err != nil {
			return ungerr.Wrapf(err, "error unmarshaling payload to: %T", taskMsg)
		}

		text, err := billSvc.ExtractBillText(ctx, taskMsg.Message)
		if err != nil {
			return err
		}

		msg := message.ExpenseBillTextExtracted{
			ID:   taskMsg.Message.ID,
			Text: text,
		}

		return extractedQueue.Enqueue(ctx, config.AppName, msg)
	})
}
