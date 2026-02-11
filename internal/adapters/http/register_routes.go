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
)

func registerRoutes(router *gin.Engine, configs config.Config, services *provider.Services, adminServices *admin.Services) {
	handlers := handler.ProvideHandlers(services)
	adminHandlers := adminHandler.ProvideHandlers(adminServices)
	middlewares := middlewares.Provide(configs.App, services.Auth)

	router.Use(middlewares.Logger, middlewares.CORS, middlewares.RateLimit, middlewares.Err)

	routes.RegisterAPIRoutes(router, handlers, middlewares)
	routes.RegisterAdminRoutes(router, adminHandlers, middlewares)
}
