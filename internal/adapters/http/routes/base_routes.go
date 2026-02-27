package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

func RegisterBaseRoutes(r *gin.Engine) {
	r.NoMethod(server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return nil, nil
		}
		return nil, ungerr.MethodNotAllowedError("method not allowed")
	}))

	r.NoRoute(server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return nil, nil
		}
		return nil, ungerr.NotFoundError("route not found")
	}))
}
