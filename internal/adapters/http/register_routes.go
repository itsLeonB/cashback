package http

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	adminHandler "github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/adapters/http/routes"
	"github.com/itsLeonB/cashback/internal/core/config"
	_ "github.com/itsLeonB/cashback/docs"
	"github.com/itsLeonB/cashback/internal/provider"
	"github.com/itsLeonB/cashback/internal/provider/admin"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/go-authkit/authgin"
	"github.com/kroma-labs/sentinel-go/httpserver"
	sentinelGin "github.com/kroma-labs/sentinel-go/httpserver/adapters/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func RegisterRoutes(router *gin.Engine, configs config.Config, services *provider.Services, adminServices *admin.Services) func() {
	authCfg := configs.Auth

	transport := authgin.NewCookieTransport(authgin.CookieConfig{
		Domain:     authCfg.CookieDomain,
		Secure:     authCfg.CookieSecure,
		SameSite:   authCfg.ParsedSameSite(),
		AccessTTL:  authCfg.TokenDuration,
		RefreshTTL: authCfg.RefreshTokenDuration,
	})

	authMW := authgin.AuthMiddleware(services.AuthKit, transport, authkit.RequireAuth)

	handlers := handler.ProvideHandlers(services, authCfg)
	adminHandlers := adminHandler.ProvideHandlers(adminServices, services)
	mw := middlewares.Provide(configs.App, adminServices.Auth)

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
	routes.RegisterAPIRoutes(router, handlers, authMW)
	routes.RegisterAdminRoutes(router, adminHandlers, mw.AdminAuth)

	return handlers.Shutdown
}
