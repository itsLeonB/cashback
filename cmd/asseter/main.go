package main

import (
	"context"

	appembed "github.com/itsLeonB/cashback"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/go-crud"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("Asseter")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	logger.Info("setting up resource for transfer methods sync...")

	svc, err := setup()
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("starting transfer methods sync...")

	if err = svc.SyncMethods(context.Background()); err != nil {
		logger.Fatal(err)
	}

	logger.Info("success syncing transfer methods")
}

func setup() (service.TransferMethodService, error) {
	dataSources, err := provider.ProvideDataSource()
	if err != nil {
		return nil, err
	}

	storageRepo, err := storage.NewGCSStorageRepository()
	if err != nil {
		return nil, err
	}

	transferMethodRepo := crud.NewRepository[debts.TransferMethod](dataSources.Gorm)
	svc := service.NewTransferMethodService(transferMethodRepo, storageRepo, "transfer-methods", appembed.TransferMethodAssets)

	return svc, nil
}
