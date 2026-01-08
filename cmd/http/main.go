package main

import (
	"github.com/itsLeonB/cashback/cmd/migrator"
	"github.com/itsLeonB/cashback/internal/adapters/http"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	logger.Init("HTTP server")

	if err := config.Load(); err != nil {
		logger.Fatal(err)
	}

	if err := migrator.Run(); err != nil {
		logger.Fatal(err)
	}

	srv, err := http.Setup(*config.Global)
	if err != nil {
		logger.Fatal(err)
	}

	srv.ServeGracefully()
}
