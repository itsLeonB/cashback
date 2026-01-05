package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/adapters/http/handler"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/provider"
)

func registerRoutes(router *gin.Engine, configs config.Config, services *provider.Services) {
	handlers := handler.ProvideHandlers(services)
	middlewares := provideMiddlewares(configs.App, services.Auth)

	router.Use(middlewares.logger, middlewares.cors, middlewares.rateLimit, middlewares.err)

	apiRoutes := router.Group("/api")
	{
		v1 := apiRoutes.Group("/v1")
		{
			authRoutes := v1.Group("/auth")
			{
				authRoutes.POST("/register", handlers.Auth.HandleRegister())
				authRoutes.POST("/login", handlers.Auth.HandleInternalLogin())
				authRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextProvider.String()), handlers.Auth.HandleOAuth2Login())
				authRoutes.GET(fmt.Sprintf("/:%s/callback", appconstant.ContextProvider.String()), handlers.Auth.HandleOAuth2Callback())
				authRoutes.GET("/verify-registration", handlers.Auth.HandleVerifyRegistration())
				authRoutes.POST("/password-reset", handlers.Auth.HandleSendPasswordReset())
				authRoutes.PATCH("/reset-password", handlers.Auth.HandleResetPassword())
			}

			protectedRoutes := v1.Group("/", middlewares.auth)
			{
				profileRoutes := protectedRoutes.Group("/profile")
				{
					profileRoutes.GET("", handlers.Profile.HandleProfile())
					profileRoutes.PATCH("", handlers.Profile.HandleUpdate())
					profileRoutes.POST("/associate", handlers.Profile.HandleAssociate())
				}

				protectedRoutes.GET("/profiles", handlers.Profile.HandleSearch())
				protectedRoutes.POST(fmt.Sprintf("/profiles/:%s/friend-requests", appconstant.ContextProfileID.String()), handlers.FriendshipRequest.HandleSend())

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

				protectedRoutes.GET("/transfer-methods", handlers.TransferMethod.HandleGetAll())
				protectedRoutes.POST("/debts", handlers.Debt.HandleCreate())
				protectedRoutes.GET("/debts", handlers.Debt.HandleGetAll())

				groupExpenseRoutes := protectedRoutes.Group("/group-expenses")
				{
					groupExpenseRoutes.POST("", handlers.GroupExpense.HandleCreateDraft())
					groupExpenseRoutes.GET("", handlers.GroupExpense.HandleGetAllCreated())
					groupExpenseRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleGetDetails())
					groupExpenseRoutes.PATCH(fmt.Sprintf("/:%s/confirmed", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleConfirmDraft())
					groupExpenseRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleDelete())
					groupExpenseRoutes.GET("/fee-calculation-methods", handlers.OtherFee.HandleGetFeeCalculationMethods())
					groupExpenseRoutes.PUT(fmt.Sprintf("/:%s/participants", appconstant.ContextGroupExpenseID.String()), handlers.GroupExpense.HandleSyncParticipants())
					groupExpenseRoutes.POST(fmt.Sprintf("/:%s/bills", appconstant.ContextGroupExpenseID.String()), handlers.ExpenseBill.HandleSave())
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
			}
		}
	}
}
