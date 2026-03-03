package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type DebtHandler struct {
	debtService service.DebtService
}

func NewDebtHandler(debtService service.DebtService) *DebtHandler {
	return &DebtHandler{debtService}
}

func (dh *DebtHandler) HandleCreate() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleCreate", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.NewDebtTransactionRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.UserProfileID = profileID

		return dh.debtService.RecordNewTransaction(ctx.Request.Context(), request)
	})
}

func (dh *DebtHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetTransactions(ctx.Request.Context(), profileID)
	})
}

func (dh *DebtHandler) HandleGetTransactionSummary() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetTransactionSummary", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetTransactionSummary(ctx.Request.Context(), profileID)
	})
}

func (dh *DebtHandler) HandleGetRecent() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetRecent", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetRecent(ctx.Request.Context(), profileID)
	})
}
