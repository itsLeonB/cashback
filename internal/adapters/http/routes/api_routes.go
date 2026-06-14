package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/kroma-labs/sentinel-go/httpserver"
	sentinelGin "github.com/kroma-labs/sentinel-go/httpserver/adapters/gin"
	"golang.org/x/time/rate"
)

func RegisterAPIRoutes(router *gin.Engine, handlers *handler.Handlers, authMiddleware gin.HandlerFunc) {
	apiRoutes := router.Group("/api")
	{
		v1 := apiRoutes.Group("/v1")
		{
			v1.POST("/payments/midtrans/notifications", handlers.Payment.HandleNotification())
			v1.GET("/plans", handlers.Plan.HandleGetActive())
			v1.GET("/public/profiles/:slug", handlers.Public.HandleGetPublicProfile())

			authRoutes := v1.Group("/auth")
			authRoutes.Use(sentinelGin.RateLimit(httpserver.RateLimitConfig{
				Limit:   rate.Limit(20.0 / 60),
				Burst:   5,
				KeyFunc: httpserver.KeyFuncByIP(),
			}))
			{
				authRoutes.POST("/register", handlers.Auth.Register())
				authRoutes.POST("/login", handlers.Auth.Login())
				authRoutes.PUT("/refresh", handlers.Auth.RefreshToken())
				authRoutes.GET("/:provider", handlers.Auth.OAuthLogin())
				authRoutes.GET("/:provider/callback", handlers.Auth.OAuthCallback())
				authRoutes.GET("/verify-registration", handlers.Auth.VerifyRegistration())
				authRoutes.POST("/password-reset",
					sentinelGin.RateLimit(httpserver.RateLimitConfig{
						Limit:   rate.Limit(3.0 / 900),
						Burst:   3,
						KeyFunc: httpserver.KeyFuncByIP(),
					}),
					handlers.Auth.SendPasswordReset(),
				)
				authRoutes.PATCH("/reset-password",
					sentinelGin.RateLimit(httpserver.RateLimitConfig{
						Limit:   rate.Limit(5.0 / 900),
						Burst:   5,
						KeyFunc: httpserver.KeyFuncByIP(),
					}),
					handlers.Auth.ResetPassword(),
				)
			}

			protectedRoutes := v1.Group("/", authMiddleware, middlewares.CSRF())
			{
				protectedRoutes.DELETE("/auth/logout", handlers.Auth.Logout())

				transferMethodsRoute := "/transfer-methods"
				profileRoutes := protectedRoutes.Group("/profile")
				{
					profileRoutes.GET("", handlers.Profile.HandleProfile())
					profileRoutes.PATCH("", handlers.Profile.HandleUpdate())
					profileRoutes.POST("/associate", handlers.Profile.HandleAssociate())
					profileRoutes.POST(transferMethodsRoute, handlers.ProfileTransferMethod.HandleAdd())
					profileRoutes.GET(transferMethodsRoute, handlers.ProfileTransferMethod.HandleGetAllOwned())
					profileRoutes.GET("/subscription", handlers.Subscription.HandleGetSubscribedDetails())
				}

				profilesRoutes := protectedRoutes.Group("/profiles")
				profilesRoutes.Use(
					middlewares.WithRateKey(appconstant.ContextProfileID.String()),
					sentinelGin.RateLimit(httpserver.RateLimitConfig{
						Limit:   rate.Limit(10.0 / 60),
						Burst:   3,
						KeyFunc: httpserver.KeyFuncByHeader("X-Rate-Key"),
					}),
				)
				{
					profilesRoutes.GET("", handlers.Profile.HandleSearch())
					profilesRoutes.POST(fmt.Sprintf("/:%s/friend-requests", appconstant.ContextProfileID.String()), handlers.FriendshipRequest.HandleSend())
					profilesRoutes.GET(fmt.Sprintf("/:%s%s", appconstant.ContextProfileID.String(), transferMethodsRoute), handlers.ProfileTransferMethod.HandleGetAllByFriendProfileID())
				}

				friendshipRoutes := protectedRoutes.Group("/friendships")
				{
					friendshipRoutes.POST("", handlers.Friendship.HandleCreateAnonymousFriendship())
					friendshipRoutes.GET("", handlers.Friendship.HandleGetAll())
					friendshipRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextFriendshipID), handlers.Friendship.HandleGetDetails())
				}

				receivedFriendRequestRoute := fmt.Sprintf("/%s/:%s", appconstant.ReceivedFriendRequest, appconstant.ContextFriendRequestID)
				friendRequestRoutes := protectedRoutes.Group("/friend-requests")
				{
					friendRequestRoutes.GET(fmt.Sprintf("/:%s", appconstant.PathFriendRequestType), handlers.FriendshipRequest.HandleGetAll())
					friendRequestRoutes.DELETE(fmt.Sprintf("/%s/:%s", appconstant.SentFriendRequest, appconstant.ContextFriendRequestID), handlers.FriendshipRequest.HandleCancel())
					friendRequestRoutes.DELETE(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleIgnore())
					friendRequestRoutes.PATCH(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleBlock())
					friendRequestRoutes.POST(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleAccept())
				}

				protectedRoutes.GET(transferMethodsRoute, handlers.TransferMethod.HandleGetAll())

				debtsRoutes := protectedRoutes.Group("/debts")
				{
					debtsRoutes.POST("", handlers.Debt.HandleCreate())
					debtsRoutes.GET("", handlers.Debt.HandleGetAll())
					debtsRoutes.GET("/summary", handlers.Debt.HandleGetTransactionSummary())
					debtsRoutes.GET("/recent", handlers.Debt.HandleGetRecent())
				}

				groupExpenseRoutes := protectedRoutes.Group("/group-expenses")
				{
					groupExpenseRoutes.POST("", handlers.GroupExpense.HandleCreateDraft())
					groupExpenseRoutes.GET("", handlers.GroupExpense.HandleGetAll())
					groupExpenseRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleGetDetails())
					groupExpenseRoutes.PATCH(fmt.Sprintf("/:%s/confirmed", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleConfirmDraft())
					groupExpenseRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleDelete())
					groupExpenseRoutes.GET("/fee-calculation-methods", handlers.OtherFee.HandleGetFeeCalculationMethods())
					groupExpenseRoutes.PUT(fmt.Sprintf("/:%s/participants", appconstant.ContextGroupExpenseID.String()), handlers.GroupExpense.HandleSyncParticipants())
					groupExpenseRoutes.POST(fmt.Sprintf("/:%s/bills", appconstant.ContextGroupExpenseID.String()), handlers.ExpenseBill.HandlePresignedSave())
					groupExpenseRoutes.PUT(fmt.Sprintf("/:%s/bills/:%s", appconstant.ContextGroupExpenseID.String(), appconstant.ContextExpenseBillID.String()), handlers.ExpenseBill.HandleTriggerParsing())
					groupExpenseRoutes.GET("/recent", handlers.GroupExpense.HandleGetRecent())
				}

				expenseItemRoutes := groupExpenseRoutes.Group(fmt.Sprintf("/:%s/items", appconstant.ContextGroupExpenseID))
				{
					expenseItemRoute := fmt.Sprintf("/:%s", appconstant.ContextExpenseItemID)
					expenseItemRoutes.POST("", handlers.ExpenseItem.HandleAdd())
					expenseItemRoutes.PUT(expenseItemRoute, handlers.ExpenseItem.HandleUpdate())
					expenseItemRoutes.DELETE(expenseItemRoute, handlers.ExpenseItem.HandleRemove())
					expenseItemRoutes.PUT(expenseItemRoute+"/participants", handlers.ExpenseItem.HandleSyncParticipants())
				}

				otherFeeRoutes := groupExpenseRoutes.Group(fmt.Sprintf("/:%s/fees", appconstant.ContextGroupExpenseID))
				{
					otherFeeRoutes.POST("", handlers.OtherFee.HandleAdd())
					otherFeeRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextOtherFeeID), handlers.OtherFee.HandleUpdate())
					otherFeeRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextOtherFeeID), handlers.OtherFee.HandleRemove())
				}

				notificationRoutes := protectedRoutes.Group("/notifications")
				{
					notificationRoutes.GET("", handlers.Notification.HandleGetUnread())
					notificationRoutes.PATCH(fmt.Sprintf("/:%s", appconstant.ContextNotificationID), handlers.Notification.HandleMarkAsRead())
					notificationRoutes.PATCH("", handlers.Notification.HandleMarkAllAsRead())
				}

				pushRoutes := protectedRoutes.Group("/push")
				{
					pushRoutes.POST("/subscribe", handlers.PushSubscription.HandleSubscribe())
					pushRoutes.POST("/unsubscribe", handlers.PushSubscription.HandleUnsubscribe())
				}

				protectedRoutes.POST(fmt.Sprintf("/plans/:%s/versions/:%s/subscriptions", appconstant.ContextPlanID.String(), appconstant.ContextPlanVersionID.String()), handlers.Subscription.HandleCreatePurchase())
				protectedRoutes.POST(fmt.Sprintf("/subscriptions/:%s", appconstant.ContextSubscriptionID.String()), handlers.Payment.HandleMakePayment())
			}
		}
	}
}
