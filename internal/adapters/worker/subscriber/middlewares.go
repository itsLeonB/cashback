package subscriber

import (
	"context"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/nats-io/nats.go/jetstream"
)

func withLogging[T queue.TaskMessage](taskType string, handler func(context.Context, T) error) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		var tmsg T
		ctx, span := otel.Tracer.Start(context.Background(), tmsg.Type())
		defer span.End()

		logger.Infof("received new task %s", taskType)

		parsed, err := ezutil.Unmarshal[T](msg.Data())
		if err != nil {
			logger.Errorf("error processing %s task: %v", taskType, err)
			_ = msg.Nak()
			return
		}

		if err := handler(ctx, parsed); err != nil {
			logger.Errorf("error processing %s task: %v", taskType, err)
			_ = msg.Nak()
			return
		}

		logger.Infof("success processing %s task", taskType)
		_ = msg.Ack()
	}
}
