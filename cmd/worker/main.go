package main

import (
	"github.com/itsLeonB/cashback/cmd/migrator"
	"github.com/itsLeonB/cashback/internal/adapters/worker"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("Worker")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	if err := migrator.Run(); err != nil {
		logger.Fatal(err)
	}

	w, err := worker.Setup()
	if err != nil {
		logger.Fatal(err)
	}

	w.Run()
}
