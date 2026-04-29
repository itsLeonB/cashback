package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type groupExpenseHandler struct {
	groupExpenseService service.GroupExpenseService
}

func newGroupExpenseHandler(
	groupExpenseService service.GroupExpenseService,
) *groupExpenseHandler {
	return &groupExpenseHandler{
		groupExpenseService,
	}
}

// HandleCreateDraft godoc
// @Summary      Create a draft group expense
// @Tags         group-expenses
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.NewDraftRequest true "New draft payload"
// @Success      201  {object}  response.JSONResponse[dto.GroupExpenseResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses [post]
func (geh *groupExpenseHandler) HandleCreateDraft() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleCreateDraft", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.NewDraftRequest](ctx)
		if err != nil {
			return nil, err
		}

		return geh.groupExpenseService.CreateDraft(ctx.Request.Context(), userProfileID, req.Description)
	})
}

// HandleGetAll godoc
// @Summary      Get all group expenses
// @Tags         group-expenses
// @Security     BearerAuth
// @Produce      json
// @Param        status    query string false "Filter by status"
// @Param        ownership query string false "Filter by ownership (default: owned)"
// @Success      200  {object}  response.JSONResponse[[]dto.GroupExpenseResponse]
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses [get]
func (geh *groupExpenseHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		status := expenses.ExpenseStatus(ctx.Query("status"))
		ownership := expenses.ExpenseOwnership(ctx.Query("ownership"))

		if ownership == "" {
			ownership = expenses.OwnedExpense
		}

		groupExpenses, err := geh.groupExpenseService.GetAll(ctx.Request.Context(), userProfileID, ownership, status)
		if err != nil {
			return nil, err
		}

		return groupExpenses, nil
	})
}

// HandleGetDetails godoc
// @Summary      Get group expense details
// @Tags         group-expenses
// @Security     BearerAuth
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Success      200  {object}  response.JSONResponse[dto.GroupExpenseResponse]
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId} [get]
func (geh *groupExpenseHandler) HandleGetDetails() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleGetDetails", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		return geh.groupExpenseService.GetDetails(ctx.Request.Context(), groupExpenseID, userProfileID)
	})
}

// HandleConfirmDraft godoc
// @Summary      Confirm a draft group expense
// @Tags         group-expenses
// @Security     BearerAuth
// @Produce      json
// @Param        groupExpenseId path  string true  "Group expense ID"
// @Param        dry-run        query string false "Set to true for dry run"
// @Success      200  {object}  response.JSONResponse[dto.ExpenseConfirmationResponse]
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/confirmed [patch]
func (geh *groupExpenseHandler) HandleConfirmDraft() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleConfirmDraft", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		var dryRun bool
		if ctx.Query("dry-run") == "true" {
			dryRun = true
		}

		return geh.groupExpenseService.ConfirmDraft(ctx.Request.Context(), groupExpenseID, userProfileID, dryRun)
	})
}

// HandleDelete godoc
// @Summary      Delete a group expense
// @Tags         group-expenses
// @Security     BearerAuth
// @Param        groupExpenseId path string true "Group expense ID"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId} [delete]
func (geh *groupExpenseHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleDelete", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.groupExpenseService.Delete(ctx.Request.Context(), userProfileID, expenseID)
	})
}

// HandleSyncParticipants godoc
// @Summary      Sync participants of a group expense
// @Tags         group-expenses
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        body body dto.ExpenseParticipantsRequest true "Participants payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/participants [put]
func (geh *groupExpenseHandler) HandleSyncParticipants() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleSyncParticipants", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.ExpenseParticipantsRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.UserProfileID = userProfileID
		req.GroupExpenseID = expenseID

		return nil, geh.groupExpenseService.SyncParticipants(ctx.Request.Context(), req)
	})
}

// HandleGetRecent godoc
// @Summary      Get recent group expenses
// @Tags         group-expenses
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.GroupExpenseResponse]
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/recent [get]
func (geh *groupExpenseHandler) HandleGetRecent() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleGetRecent", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return geh.groupExpenseService.GetRecent(ctx.Request.Context(), profileID)
	})
}
