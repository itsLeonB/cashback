package logger

import (
	"context"
	"io"
	"time"

	"go.opentelemetry.io/otel/log"
)

var _ io.Writer = (*otelWriter)(nil)

type otelWriter struct {
	logger log.Logger
}

func (w *otelWriter) Write(p []byte) (n int, err error) {
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetBody(log.StringValue(string(p)))
	w.logger.Emit(context.Background(), record)
	return len(p), nil
}
