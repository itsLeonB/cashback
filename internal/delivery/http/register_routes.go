package http

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/config"
	"github.com/itsLeonB/orcashtrator/internal/delivery/http/handler"
	"github.com/itsLeonB/orcashtrator/internal/provider"
)

func registerRoutes(router *gin.Engine, configs config.Config, logger ezutil.Logger, services *provider.Services) {
	handlers := handler.ProvideHandlers(logger, services)
	middlewares := provideMiddlewares(configs.App, logger, services.Auth)

	router.Use(middlewares.logger, middlewares.cors)

	apiRoutes := router.Group("/api", middlewares.err)

	v1 := apiRoutes.Group("/v1")

	authRoutes := v1.Group("/auth")
	authRoutes.POST("/register", handlers.Auth.HandleRegister())
	authRoutes.POST("/login", handlers.Auth.HandleInternalLogin())
	authRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextProvider), handlers.Auth.HandleOAuth2Login())
	authRoutes.GET(fmt.Sprintf("/:%s/callback", appconstant.ContextProvider), handlers.Auth.HandleOAuth2Callback())
	authRoutes.GET("/verify-registration", handlers.Auth.HandleVerifyRegistration())
	authRoutes.POST("/password-reset", handlers.Auth.HandleSendPasswordReset())
	authRoutes.PATCH("/reset-password", handlers.Auth.HandleResetPassword())

	protectedRoutes := v1.Group("/", middlewares.auth)

	profileRoutes := protectedRoutes.Group("/profile")
	profileRoutes.GET("", handlers.Profile.HandleProfile())
	profileRoutes.PATCH("", handlers.Profile.HandleUpdate())

	protectedRoutes.GET("/profiles", handlers.Profile.HandleSearch())
	protectedRoutes.POST(fmt.Sprintf("/profiles/:%s/friend-requests", appconstant.ContextProfileID.String()), handlers.FriendshipRequest.HandleSend())

	friendshipRoutes := protectedRoutes.Group("/friendships")
	friendshipRoutes.POST("", handlers.Friendship.HandleCreateAnonymousFriendship())
	friendshipRoutes.GET("", handlers.Friendship.HandleGetAll())
	friendshipRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextFriendshipID), handlers.Friendship.HandleGetDetails())

	receivedFriendRequestRoute := fmt.Sprintf("/%s/:%s", appconstant.ReceivedFriendRequest, appconstant.ContextFriendRequestID)
	friendRequestRoutes := protectedRoutes.Group("/friend-requests")
	friendRequestRoutes.GET(fmt.Sprintf("/:%s", appconstant.PathFriendRequestType), handlers.FriendshipRequest.HandleGetAll())
	friendRequestRoutes.DELETE(fmt.Sprintf("/%s/:%s", appconstant.SentFriendRequest, appconstant.ContextFriendRequestID), handlers.FriendshipRequest.HandleCancel())
	friendRequestRoutes.DELETE(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleIgnore())
	friendRequestRoutes.PATCH(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleBlock())
	friendRequestRoutes.POST(receivedFriendRequestRoute, handlers.FriendshipRequest.HandleAccept())

	protectedRoutes.GET("/transfer-methods", handlers.TransferMethod.HandleGetAll())

	protectedRoutes.POST("/debts", handlers.Debt.HandleCreate())
	protectedRoutes.GET("/debts", handlers.Debt.HandleGetAll())

	groupExpenseRoutes := protectedRoutes.Group("/group-expenses")
	groupExpenseRoutes.POST("", handlers.GroupExpense.HandleCreateDraft())
	groupExpenseRoutes.GET("", handlers.GroupExpense.HandleGetAllCreated())
	groupExpenseRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleGetDetails())
	groupExpenseRoutes.PATCH(fmt.Sprintf("/:%s/confirmed", appconstant.ContextGroupExpenseID), handlers.GroupExpense.HandleConfirmDraft())

	expenseItemRoute := fmt.Sprintf("/:%s", appconstant.ContextExpenseItemID)
	expenseItemRoutes := groupExpenseRoutes.Group(fmt.Sprintf("/:%s/items", appconstant.ContextGroupExpenseID))
	expenseItemRoutes.POST("", handlers.ExpenseItem.HandleAdd())
	expenseItemRoutes.GET(expenseItemRoute, handlers.ExpenseItem.HandleGetDetails())
	expenseItemRoutes.PUT(expenseItemRoute, handlers.ExpenseItem.HandleUpdate())
	expenseItemRoutes.DELETE(expenseItemRoute, handlers.ExpenseItem.HandleRemove())

	otherFeeRoutes := groupExpenseRoutes.Group(fmt.Sprintf("/:%s/fees", appconstant.ContextGroupExpenseID))
	otherFeeRoutes.POST("", handlers.OtherFee.HandleAdd())
	otherFeeRoutes.PUT(fmt.Sprintf("/:%s", appconstant.ContextOtherFeeID), handlers.OtherFee.HandleUpdate())
	otherFeeRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextOtherFeeID), handlers.OtherFee.HandleRemove())

	groupExpenseRoutes.GET("/fee-calculation-methods", handlers.OtherFee.HandleGetFeeCalculationMethods())

	expenseBillRoutes := groupExpenseRoutes.Group("/bills")
	expenseBillRoutes.POST("", handlers.ExpenseBill.HandleSave())
	expenseBillRoutes.GET("", handlers.ExpenseBill.HandleGetAllCreated())
	expenseBillRoutes.GET(fmt.Sprintf("/:%s", appconstant.ContextExpenseBillID.String()), handlers.ExpenseBill.HandleGet())
	expenseBillRoutes.DELETE(fmt.Sprintf("/:%s", appconstant.ContextExpenseBillID.String()), handlers.ExpenseBill.HandleDelete())
}
