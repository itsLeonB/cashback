package http

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

func Setup(configs config.Config) (*server.Http, error) {
	providers, err := provider.All()
	if err != nil {
		return nil, err
	}

	gin.SetMode(configs.Env)
	r := gin.New()
	registerRoutes(r, configs, providers.Services, providers.AdminServices)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", configs.App.Port),
		Handler:           r,
		ReadTimeout:       configs.Timeout,
		ReadHeaderTimeout: configs.Timeout,
		WriteTimeout:      configs.Timeout,
		IdleTimeout:       configs.Timeout,
	}

	return server.New(
		srv,
		configs.Timeout,
		logger.Global,
		providers.Shutdown,
	), nil
}
