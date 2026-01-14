package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type ExpenseItemHandler struct {
	expenseItemSvc service.ExpenseItemService
}

func NewExpenseItemHandler(
	expenseItemSvc service.ExpenseItemService,
) *ExpenseItemHandler {
	return &ExpenseItemHandler{
		expenseItemSvc,
	}
}

func (geh *ExpenseItemHandler) HandleAdd() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.NewExpenseItemRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID

		return nil, geh.expenseItemSvc.Add(ctx, request)
	})
}

func (geh *ExpenseItemHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		expenseItemID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.UpdateExpenseItemRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID
		request.ID = expenseItemID

		return nil, geh.expenseItemSvc.Update(ctx, request)
	})
}

func (geh *ExpenseItemHandler) HandleRemove() gin.HandlerFunc {
	return server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		expenseItemID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.expenseItemSvc.Remove(ctx, groupExpenseID, expenseItemID, userProfileID)
	})
}

func (geh *ExpenseItemHandler) HandleSyncParticipants() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		expenseItemID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.SyncItemParticipantsRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.ProfileID = userProfileID
		request.ID = expenseItemID
		request.GroupExpenseID = groupExpenseID

		return nil, geh.expenseItemSvc.SyncParticipants(ctx, request)
	})
}
