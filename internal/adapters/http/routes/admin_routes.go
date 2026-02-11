package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
)

func RegisterAdminRoutes(router *gin.Engine, handlers *admin.Handlers, middlewares *middlewares.Middlewares) {
	adminRoutes := router.Group("/admin")
	{
		v1 := adminRoutes.Group("/v1")
		{
			authRoutes := v1.Group("/auth")
			{
				authRoutes.POST("/register", handlers.Auth.HandleRegister())
				authRoutes.POST("/login", handlers.Auth.HandleLogin())
			}
		}
	}
}
