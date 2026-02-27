package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

func RegisterTestRoutes(r *gin.Engine) {
	r.GET("/success", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, "created")
	})
	r.GET("/error", func(ctx *gin.Context) {
		ctx.JSON(http.StatusInternalServerError, "error")
	})
	r.POST("/post-error", func(ctx *gin.Context) {
		ctx.JSON(http.StatusBadRequest, "error post")
	})
	group := r.Group("/test")
	{
		group.GET("/success", handleSuccess())
		group.GET("/error", handleError())
		group.GET("/wrapped-error", handleWrappedError())
		group.GET("/unwrapped-error", handleUnwrappedError())
		group.GET("/app-error", handleAppError())
		group.GET("/known-error", handleKnownError())
		group.GET("/panic", handlePanic())
	}
}

func handleSuccess() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return "success", nil
	})
}

func handleError() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.Unknown("this error should be handled as InternalServerError")
	})
}

func handleWrappedError() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.Wrap(http.ErrNoCookie, "no cookie")
	})
}

func handleUnwrappedError() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, http.ErrNoCookie
	})
}

func handleAppError() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, ungerr.BadRequestError("error should be returned")
	})
}

func handleKnownError() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return nil, &json.SyntaxError{}
	})
}

func handlePanic() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		panic("panicking")
	})
}
