package main

import (
	"github.com/itsLeonB/cashback/internal/adapters/job"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("Job")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	j, err := job.Setup(config.Global)
	if err != nil {
		logger.Fatal(err)
	}

	j.Run()
}
