package subscriber

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/codes"
)

func withLogging[T queue.TaskMessage](taskType string, handler func(context.Context, T) error) jetstream.MessageHandler {
	return func(msg jetstream.Msg) {
		logger.Infof("received new task %s", taskType)

		ctx, span := otel.Tracer.Start(context.Background(), taskType)
		defer span.End()

		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				err := fmt.Errorf("panic in task %s: %v\n%s", taskType, r, stack)
				span.RecordError(err)
				span.SetStatus(codes.Error, "panic recovered")
				logger.Errorf("panic in task %s: %v\n%s", taskType, r, stack)
				_ = msg.Nak()
			}
		}()

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
