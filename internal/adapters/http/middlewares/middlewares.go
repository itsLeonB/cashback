package middlewares

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/admin"
	"github.com/itsLeonB/ginkgo/pkg/middleware"
	"golang.org/x/time/rate"
)

type Middlewares struct {
	Auth      gin.HandlerFunc
	Err       gin.HandlerFunc
	CORS      gin.HandlerFunc
	Logger    gin.HandlerFunc
	RateLimit gin.HandlerFunc
	AdminAuth gin.HandlerFunc
}

func Provide(configs config.App, authSvc service.AuthService, adminAuthSvc admin.AuthService) *Middlewares {
	tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		return authSvc.VerifyToken(ctx, token)
	}

	adminTokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		return adminAuthSvc.VerifyToken(ctx, token)
	}

	middlewareProvider := middleware.NewMiddlewareProvider(logger.Global)
	authMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", tokenCheckFunc)
	adminAuthMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", adminTokenCheckFunc)
	errorMiddleware := middlewareProvider.NewErrorMiddleware()
	loggingMiddleware := middlewareProvider.NewLoggingMiddleware()
	rateLimitMiddleware := middlewareProvider.NewRateLimitMiddleware(10*rate.Every(time.Second), 10)

	corsMiddleware := middlewareProvider.NewCorsMiddleware(&cors.Config{
		AllowOrigins: configs.ClientUrls,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Origin",
			"Cache-Control",
			"Referer",
			"User-Agent",
			"range",
			"DNT",
			"sec-ch-ua",
			"sec-ch-ua-platform",
			"sec-ch-ua-mobile",
		},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count"},
		AllowCredentials: true,
	})

	return &Middlewares{
		authMiddleware,
		errorMiddleware,
		corsMiddleware,
		loggingMiddleware,
		rateLimitMiddleware,
		adminAuthMiddleware,
	}
}
