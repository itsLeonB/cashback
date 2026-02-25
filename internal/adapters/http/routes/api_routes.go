package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	"github.com/itsLeonB/cashback/internal/adapters/http/middlewares"
	"github.com/itsLeonB/cashback/internal/appconstant"
)

func RegisterAPIRoutes(router *gin.Engine, handlers *handler.Handlers, middlewares *middlewares.Middlewares) {
	apiRoutes := router.Group("/api")
	{
		v1 := apiRoutes.Group("/v1")
		{
			v1.POST("/payments/midtrans/notifications", handlers.Payment.HandleNotification())

			authRoutes := v1.Group("/auth")
			{
				authRoutes.POST("/register", handlers.Auth.HandleRegister())
				authRoutes.POST("/login", handlers.Auth.HandleInternalLogin())
				authRoutes.PUT("/refresh", handlers.Auth.HandleRefreshToken())
				authRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextProvider.String()), handlers.Auth.HandleOAuth2Login())
				authRoutes.GET(fmt.Sprintf("/:%s/callback", appconstant.ContextProvider.String()), handlers.Auth.HandleOAuth2Callback())
				authRoutes.GET("/verify-registration", handlers.Auth.HandleVerifyRegistration())
				authRoutes.POST("/password-reset", handlers.Auth.HandleSendPasswordReset())
				authRoutes.PATCH("/reset-password", handlers.Auth.HandleResetPassword())
			}

			protectedRoutes := v1.Group("/", middlewares.Auth)
			{
				protectedRoutes.DELETE("/auth/logout", handlers.Auth.HandleLogout())

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
					groupExpenseRoutes.POST(fmt.Sprintf("/:%s/bills", appconstant.ContextGroupExpenseID.String()), handlers.ExpenseBill.HandleSave())
					groupExpenseRoutes.PUT(fmt.Sprintf("/:%s/bills/%s", appconstant.ContextGroupExpenseID.String(), appconstant.ContextExpenseBillID.String()), handlers.ExpenseBill.HandleTriggerParsing())
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

				planRoutes := protectedRoutes.Group("/plans")
				{
					planRoutes.POST(fmt.Sprintf("/:%s/versions/:%s/subscriptions", appconstant.ContextPlanID.String(), appconstant.ContextPlanVersionID.String()), handlers.Subscription.HandleCreatePurchase())
					planRoutes.GET("", handlers.Plan.HandleGetActive())
				}
			}
		}
	}
}
