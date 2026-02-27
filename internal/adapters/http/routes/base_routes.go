package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

func RegisterBaseRoutes(r *gin.Engine) {
	r.NoMethod(server.Handler(http.StatusMethodNotAllowed, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.MethodNotAllowedError("method not allowed")
	}))

	r.NoRoute(server.Handler(http.StatusNotFound, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.NotFoundError("route not found")
	}))
}
