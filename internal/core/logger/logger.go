package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/global"
)

var Global zerolog.Logger

func Init(appNamespace string) {
	otelLogger := global.Logger(appNamespace)

	Global = zerolog.New(os.Stdout).With().Timestamp().Logger().Hook(
		zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
			record := log.Record{}
			record.SetTimestamp(time.Now())
			record.SetSeverity(mapSeverity(level))
			record.SetSeverityText(level.String())
			record.SetBody(log.StringValue(msg))
			otelLogger.Emit(context.Background(), record)
		}),
	)
}

func Debug(args ...any) {
	Global.Debug().Str("", "").Msg(fmt.Sprint(args...))
}

func Info(args ...any) {
	Global.Info().Str("", "").Msg(fmt.Sprint(args...))
}

func Warn(args ...any) {
	Global.Warn().Str("", "").Msg(fmt.Sprint(args...))
}

func Error(args ...any) {
	Global.Error().Str("", "").Msg(fmt.Sprint(args...))
}

func Fatal(args ...any) {
	Global.Fatal().Str("", "").Msg(fmt.Sprint(args...))
}

func Debugf(format string, args ...any) {
	Global.Debug().Str("", "").Msg(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	Global.Info().Str("", "").Msg(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	Global.Warn().Str("", "").Msg(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	Global.Error().Str("", "").Msg(fmt.Sprintf(format, args...))
}

func Fatalf(format string, args ...any) {
	Global.Fatal().Str("", "").Msg(fmt.Sprintf(format, args...))
}

func mapSeverity(level zerolog.Level) log.Severity {
	switch level {
	case zerolog.TraceLevel:
		return log.SeverityTrace
	case zerolog.DebugLevel:
		return log.SeverityDebug
	case zerolog.InfoLevel:
		return log.SeverityInfo
	case zerolog.WarnLevel:
		return log.SeverityWarn
	case zerolog.ErrorLevel:
		return log.SeverityError
	case zerolog.FatalLevel, zerolog.PanicLevel:
		return log.SeverityFatal
	default:
		return log.SeverityUndefined
	}
}
