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

// HandleAdd godoc
// @Summary      Add an item to a group expense
// @Tags         expense-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        body body dto.NewExpenseItemRequest true "New expense item payload"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/items [post]
func (geh *ExpenseItemHandler) HandleAdd() gin.HandlerFunc {
	return server.Handler("ExpenseItemHandler.HandleAdd", http.StatusCreated, func(ctx *gin.Context) (any, error) {
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

		return nil, geh.expenseItemSvc.Add(ctx.Request.Context(), request)
	})
}

// HandleUpdate godoc
// @Summary      Update an expense item
// @Tags         expense-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        expenseItemId  path string true "Expense item ID"
// @Param        body body dto.UpdateExpenseItemRequest true "Update expense item payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/items/{expenseItemId} [put]
func (geh *ExpenseItemHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler("ExpenseItemHandler.HandleUpdate", http.StatusOK, func(ctx *gin.Context) (any, error) {
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

		return nil, geh.expenseItemSvc.Update(ctx.Request.Context(), request)
	})
}

// HandleRemove godoc
// @Summary      Remove an expense item
// @Tags         expense-items
// @Security     BearerAuth
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        expenseItemId  path string true "Expense item ID"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/items/{expenseItemId} [delete]
func (geh *ExpenseItemHandler) HandleRemove() gin.HandlerFunc {
	return server.Handler("ExpenseItemHandler.HandleRemove", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
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

		return nil, geh.expenseItemSvc.Remove(ctx.Request.Context(), groupExpenseID, expenseItemID, userProfileID)
	})
}

// HandleSyncParticipants godoc
// @Summary      Sync participants of an expense item
// @Tags         expense-items
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        expenseItemId  path string true "Expense item ID"
// @Param        body body dto.SyncItemParticipantsRequest true "Participants payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/items/{expenseItemId}/participants [put]
func (geh *ExpenseItemHandler) HandleSyncParticipants() gin.HandlerFunc {
	return server.Handler("ExpenseItemHandler.HandleSyncParticipants", http.StatusOK, func(ctx *gin.Context) (any, error) {
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

		return nil, geh.expenseItemSvc.SyncParticipants(ctx.Request.Context(), request)
	})
}
