package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/service"
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

func (geh *groupExpenseHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		status := expenses.ExpenseStatus(ctx.Query("status"))
		ownership := expenses.ExpenseOwnership(ctx.Query("ownership"))

		// Default to OWNED if ownership not specified (backward compatibility)
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

func (geh *groupExpenseHandler) HandleGetRecent() gin.HandlerFunc {
	return server.Handler("GroupExpenseHandler.HandleGetRecent", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return geh.groupExpenseService.GetRecent(ctx.Request.Context(), profileID)
	})
}
