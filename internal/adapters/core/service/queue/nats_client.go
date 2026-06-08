package queue

import (
	"context"
	"encoding/json"
	"time"

	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/ungerr"
	"github.com/nats-io/nats.go/jetstream"
	"go.opentelemetry.io/otel/trace"
)

type natsClient struct {
	js jetstream.JetStream
}

func NewNATSTaskQueue(js jetstream.JetStream) *natsClient {
	return &natsClient{js: js}
}

func (nc *natsClient) Enqueue(ctx context.Context, message queue.TaskMessage) error {
	ctx, span := otel.Tracer.Start(ctx, "natsClient.Enqueue")
	defer span.End()

	payload, err := json.Marshal(message)
	if err != nil {
		return ungerr.Wrap(err, "error marshaling message to JSON")
	}

	ack, err := nc.js.Publish(ctx, message.Type(), payload)
	if err != nil {
		return ungerr.Wrap(err, "error publishing message to NATS")
	}

	logger.Infof("published message: Stream=%s, Seq=%d, Subject=%s", ack.Stream, ack.Sequence, message.Type())
	return nil
}

func (nc *natsClient) Shutdown() error {
	return nil
}

func (nc *natsClient) AsyncEnqueue(ctx context.Context, msg queue.TaskMessage) {
	parentSpanCtx := trace.SpanContextFromContext(ctx)
	go func() {
		detached, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if parentSpanCtx.IsValid() {
			detached = trace.ContextWithSpanContext(detached, parentSpanCtx)
		}

		if err := nc.Enqueue(detached, msg); err != nil {
			logger.Error(err)
		}
	}()
}
