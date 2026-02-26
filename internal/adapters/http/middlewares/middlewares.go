package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/cashback/internal/domain/service/admin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo/pkg/middleware"
)

type Middlewares struct {
	Auth      gin.HandlerFunc
	Err       gin.HandlerFunc
	AdminAuth gin.HandlerFunc
}

func Provide(configs config.App, authSvc service.AuthService, adminAuthSvc admin.AuthService) *Middlewares {
	tokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		return authSvc.VerifyToken(ctx, token)
	}

	adminTokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		return adminAuthSvc.VerifyToken(ctx, token)
	}

	middlewareProvider := middleware.NewMiddlewareProvider(ezutil.NewSimpleLogger(config.Global.ServiceName, true, 0))
	authMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", tokenCheckFunc)
	adminAuthMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", adminTokenCheckFunc)
	errorMiddleware := middlewareProvider.NewErrorMiddleware()

	return &Middlewares{
		authMiddleware,
		errorMiddleware,
		adminAuthMiddleware,
	}
}
