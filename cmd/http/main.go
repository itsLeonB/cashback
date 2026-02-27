package main

import (
	"context"

	"github.com/itsLeonB/cashback/internal/adapters/http"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("Cashback")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()
	otelShutdown, err := otel.InitSDK(ctx, config.Global.OTel)
	if err != nil {
		logger.Fatal(err)
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			logger.Error(err)
		}
	}()

	srv, shutdownFunc, err := http.Setup(*config.Global)
	if err != nil {
		logger.Fatal(err)
	}

	defer shutdownFunc()

	if err := srv.ListenAndServe(ctx); err != nil {
		logger.Error(err)
	}
}
