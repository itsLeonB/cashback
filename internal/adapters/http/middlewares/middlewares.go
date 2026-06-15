package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/go-authkit"
	"github.com/itsLeonB/ginkgo/pkg/middleware"
)

type Middlewares struct {
	Err       gin.HandlerFunc
	AdminAuth gin.HandlerFunc
}

func Provide(configs config.App, adminKit *authkit.AuthKit) *Middlewares {
	adminTokenCheckFunc := func(ctx *gin.Context, token string) (bool, map[string]any, error) {
		claims, err := adminKit.VerifyToken(ctx.Request.Context(), token, "")
		if err != nil {
			return false, nil, err
		}
		return true, claims, nil
	}

	middlewareProvider := middleware.NewMiddlewareProvider(logger.Global)
	adminAuthMiddleware := middlewareProvider.NewAuthMiddleware("Bearer", adminTokenCheckFunc)
	errorMiddleware := middlewareProvider.NewErrorMiddleware()

	return &Middlewares{
		errorMiddleware,
		adminAuthMiddleware,
	}
}
