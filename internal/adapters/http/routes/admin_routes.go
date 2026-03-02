package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler/admin"
	"github.com/itsLeonB/cashback/internal/appconstant"
)

func RegisterAdminRoutes(router *gin.Engine, handlers *admin.Handlers, authMiddleware gin.HandlerFunc) {
	adminRoutes := router.Group("/admin")
	{
		v1 := adminRoutes.Group("/v1")
		{
			authRoutes := v1.Group("/auth")
			{
				authRoutes.POST("/register", handlers.Auth.HandleRegister())
				authRoutes.POST("/login", handlers.Auth.HandleLogin())
			}

			protectedRoutes := v1.Group("/", authMiddleware)
			{
				protectedRoutes.GET("/auth/me", handlers.Auth.HandleMe())

				planRoutes := protectedRoutes.Group("/plans")
				{
					planRoutes.POST("", handlers.Plan.HandleCreate())
					planRoutes.GET("", handlers.Plan.HandleGetList())
					planRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleGetOne())
					planRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleUpdate())
					planRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextPlanID.String()), handlers.Plan.HandleDelete())
				}

				planVersionRoutes := protectedRoutes.Group("/plan-versions")
				{
					planVersionRoutes.POST("", handlers.PlanVersion.HandleCreate())
					planVersionRoutes.GET("", handlers.PlanVersion.HandleGetList())
					planVersionRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextPlanVersionID.String()), handlers.PlanVersion.HandleGetOne())
					planVersionRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextPlanVersionID.String()), handlers.PlanVersion.HandleUpdate())
					planVersionRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextPlanVersionID.String()), handlers.PlanVersion.HandleDelete())
				}

				subscriptionRoutes := protectedRoutes.Group("/subscriptions")
				{
					subscriptionRoutes.POST("", handlers.Subscription.HandleCreate())
					subscriptionRoutes.GET("", handlers.Subscription.HandleGetList())
					subscriptionRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextSubscriptionID.String()), handlers.Subscription.HandleGetOne())
					subscriptionRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextSubscriptionID.String()), handlers.Subscription.HandleUpdate())
					subscriptionRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextSubscriptionID.String()), handlers.Subscription.HandleDelete())
				}

				paymentRoutes := protectedRoutes.Group("/payments")
				{
					paymentRoutes.GET("", handlers.Payment.HandleGetList())
					paymentRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextPaymentID.String()), handlers.Payment.HandleGetOne())
					paymentRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextPaymentID.String()), handlers.Payment.HandleUpdate())
					paymentRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextPaymentID.String()), handlers.Payment.HandleDelete())
				}

				profileRoutes := protectedRoutes.Group("/profiles")
				{
					profileRoutes.GET("", handlers.Profile.HandleGetList())
					profileRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextProfileID.String()), handlers.Profile.HandleGetOne())
				}
			}
		}
	}
}
