package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/kroma-labs/sentinel-go/httpserver"
)

func Setup(configs config.Config) (*httpserver.Server, func(), error) {
	providers, err := provider.All()
	if err != nil {
		return nil, nil, err
	}

	shutdownFunc := func() {
		if err := providers.Shutdown(); err != nil {
			logger.Error(err)
		}
	}

	gin.SetMode(configs.App.Env)
	r := gin.New()

	skipPaths := []string{"/ping", "/livez", "/readyz", "/metrics"}
	if err = setupSentinel(r, skipPaths); err != nil {
		return nil, nil, err
	}

	RegisterRoutes(r, configs, providers.Services, providers.AdminServices)

	httpCfg := httpserver.ProductionConfig()
	httpCfg.LoggerConfig = &httpserver.LoggerConfig{
		Logger:    logger.Global,
		SkipPaths: skipPaths,
	}
	httpCfg.Addr = fmt.Sprintf(":%s", configs.App.Port)

	srv := httpserver.New(
		httpserver.WithConfig(httpCfg),
		httpserver.WithServiceName(configs.ServiceName),
		httpserver.WithHandler(r),
		httpserver.WithLogger(logger.Global),
	)

	return srv, shutdownFunc, nil
}
