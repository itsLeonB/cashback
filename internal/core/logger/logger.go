package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/log/global"
)

var Global zerolog.Logger

func Init(appNamespace string) {
	otelLogger := global.Logger(appNamespace)

	writer := zerolog.MultiLevelWriter(
		os.Stdout,
		&otelWriter{logger: otelLogger},
	)

	Global = zerolog.New(writer).With().Timestamp().Logger()
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
