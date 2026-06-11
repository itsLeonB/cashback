package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/cookie"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ungerr"
)

func newCookieAuthMiddleware(authSvc service.AuthService) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenStr, err := ctx.Cookie(cookie.AccessTokenName)
		if err != nil {
			_ = ctx.Error(ungerr.UnauthorizedError("missing access token"))
			ctx.Abort()
			return
		}

		fgp, _ := ctx.Cookie(cookie.FingerprintName)

		exists, data, err := authSvc.VerifyToken(ctx.Request.Context(), tokenStr, fgp)
		if err != nil {
			_ = ctx.Error(err)
			ctx.Abort()
			return
		}
		if !exists {
			_ = ctx.Error(ungerr.UnauthorizedError("user data not found"))
			ctx.Abort()
			return
		}

		for key, val := range data {
			ctx.Set(key, val)
		}

		ctx.Next()
	}
}
