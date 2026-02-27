package http

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	adminHandler "github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/adapters/http/routes"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/cashback/internal/provider/admin"
	"github.com/kroma-labs/sentinel-go/httpserver"
	sentinelGin "github.com/kroma-labs/sentinel-go/httpserver/adapters/gin"
)

func RegisterRoutes(router *gin.Engine, configs config.Config, services *provider.Services, adminServices *admin.Services) {
	handlers := handler.ProvideHandlers(services)
	adminHandlers := adminHandler.ProvideHandlers(adminServices, services)
	middlewares := middlewares.Provide(configs.App, services.Auth, adminServices.Auth)

	router.Use(middlewares.Err)

	sentinelGin.RegisterHealth(router, httpserver.NewHealthHandler())

	if configs.App.Env != "release" {
		sentinelGin.RegisterPprof(router, httpserver.DefaultPprofConfig())
	}

	routes.RegisterAPIRoutes(router, handlers, middlewares.Auth)
	routes.RegisterAdminRoutes(router, adminHandlers, middlewares.AdminAuth)
}
