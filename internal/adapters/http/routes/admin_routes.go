package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/appconstant"
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

			protectedRoutes := v1.Group("/", middlewares.AdminAuth)
			{
				planRoutes := protectedRoutes.Group("/plans")
				{
					planRoutes.POST("", handlers.Plan.HandleCreate())
					planRoutes.GET("", handlers.Plan.HandleGetList())
					planRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleGetOne())
					planRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleUpdate())
					planRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleDelete())
				}
			}
		}
	}
}
