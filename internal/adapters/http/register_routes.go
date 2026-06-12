package http

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/cookie"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	adminHandler "github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/adapters/http/routes"
	"github.com/itsLeonB/cashback/internal/core/config"
	_ "github.com/itsLeonB/cashback/docs"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/cashback/internal/provider/admin"
	"github.com/kroma-labs/sentinel-go/httpserver"
	sentinelGin "github.com/kroma-labs/sentinel-go/httpserver/adapters/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, configs config.Config, services *provider.Services, adminServices *admin.Services) func() {
	cookieCfg := cookie.Config{
		Domain:     configs.CookieDomain,
		Secure:     configs.CookieSecure,
		SameSite:   configs.ParsedSameSite(),
		AccessTTL:  configs.TokenDuration,
		RefreshTTL: configs.RefreshTokenDuration,
	}

	handlers := handler.ProvideHandlers(services, cookieCfg)
	adminHandlers := adminHandler.ProvideHandlers(adminServices, services)
	mw := middlewares.Provide(configs.App, services.Auth, adminServices.Auth)

	router.Use(mw.Err)

	sentinelGin.RegisterHealth(router, httpserver.NewHealthHandler())

	if configs.App.Env != "release" {
		sentinelGin.RegisterPprof(router, httpserver.DefaultPprofConfig())
		routes.RegisterTestRoutes(router)
	}

	// Swagger UI: /docs/index.html
	router.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Markdown docs: /docs.md
	router.GET("/docs.md", func(ctx *gin.Context) {
		data, err := os.ReadFile("docs/docs.md")
		if err != nil {
			ctx.Status(404)
			return
		}
		ctx.Data(200, "text/markdown; charset=utf-8", data)
	})

	routes.RegisterBaseRoutes(router)
	routes.RegisterAPIRoutes(router, handlers, mw.Auth)
	routes.RegisterAdminRoutes(router, adminHandlers, mw.AdminAuth)

	return handlers.Shutdown
}
