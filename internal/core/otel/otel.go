package otel

import (
	"context"
	"errors"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/ungerr"
	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/contrib/propagators/autoprop"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const scopeName = "github.com/itsLeonB/cashback"

var Tracer trace.Tracer = noop.NewTracerProvider().Tracer("")

// InitSDK initializes the OpenTelemetry SDK for metrics, logs, and traces.
func InitSDK(ctx context.Context, cfg config.OTel) (func(context.Context) error, error) {
	if !cfg.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	var shutdownFuncs []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		if err != nil {
			return ungerr.Wrap(err, "error shutting down otel resources")
		}
		return nil
	}

	otel.SetTextMapPropagator(autoprop.NewTextMapPropagator())

	for _, initFn := range []func(context.Context) (func(context.Context) error, error){
		initMetrics, initLogs, initTraces,
	} {
		shutdownFunc, err := initFn(ctx)
		if err != nil {
			if e := shutdown(ctx); e != nil {
				logger.Error(e)
			}
			return nil, err
		}
		shutdownFuncs = append(shutdownFuncs, shutdownFunc)
	}

	return shutdown, nil
}

func initMetrics(ctx context.Context) (func(context.Context) error, error) {
	reader, err := autoexport.NewMetricReader(ctx)
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to create metric reader")
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	otel.SetMeterProvider(mp)

	return mp.Shutdown, nil
}

func initLogs(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := autoexport.NewLogExporter(ctx)
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to create log exporter")
	}

	lp := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))
	global.SetLoggerProvider(lp)

	return lp.Shutdown, nil
}

func initTraces(ctx context.Context) (func(context.Context) error, error) {
	exporter, err := autoexport.NewSpanExporter(ctx)
	if err != nil {
		return nil, ungerr.Wrap(err, "failed to create trace exporter")
	}

	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exporter))
	otel.SetTracerProvider(tp)
	Tracer = tp.Tracer(scopeName)

	return tp.Shutdown, nil
}
