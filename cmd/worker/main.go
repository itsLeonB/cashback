package main

import (
	"context"

	"github.com/itsLeonB/cashback/internal/adapters/worker"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("Worker")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	ctx := context.Background()
	otelShutdown, err := otel.InitSDK(ctx, config.Global.OTel)
	if err != nil {
		logger.Fatalf("failed to initialize OTel SDK: %v", err)
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			logger.Errorf("error shutting down OTel SDK: %v", err)
		}
	}()

	w, err := worker.Setup()
	if err != nil {
		logger.Fatal(err)
	}

	w.Run()
}
