package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/domain/service/admin"
	"github.com/itsLeonB/ginkgo/pkg/middleware"
)

type Middlewares struct {
	Err       gin.HandlerFunc
	AdminAuth gin.HandlerFunc
}

func Provide(configs config.App, adminAuthSvc admin.AuthService) *Middlewares {
	adminTokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		return adminAuthSvc.VerifyToken(ctx.Request.Context(), token)
	}

	middlewareProvider := middleware.NewMiddlewareProvider(logger.Global)
	adminAuthMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", adminTokenCheckFunc)
	errorMiddleware := middlewareProvider.NewErrorMiddleware()

	return &Middlewares{
		errorMiddleware,
		adminAuthMiddleware,
	}
}
