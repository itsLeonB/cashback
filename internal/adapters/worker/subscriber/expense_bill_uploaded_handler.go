package subscriber

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/meq"
)

func expenseBillUploadedHandler(
	billSvc service.ExpenseBillService,
	extractedQueue meq.TaskQueue[message.ExpenseBillTextExtracted],
) asynq.Handler {
	return withLogging(message.ExpenseBillUploaded{}.Type(), func(ctx context.Context, msg message.ExpenseBillUploaded) error {
		text, err := billSvc.ExtractBillText(ctx, msg)
		if err != nil {
			return err
		}

		return extractedQueue.Enqueue(ctx, config.AppName, message.ExpenseBillTextExtracted{
			ID:   msg.ID,
			Text: text,
		})
	})
}
