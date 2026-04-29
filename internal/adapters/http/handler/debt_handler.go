package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type DebtHandler struct {
	debtService service.DebtService
}

func NewDebtHandler(debtService service.DebtService) *DebtHandler {
	return &DebtHandler{debtService}
}

// HandleCreate godoc
// @Summary      Record a new debt transaction
// @Tags         debts
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.NewDebtTransactionRequest true "New debt transaction payload"
// @Success      201  {object}  response.JSONResponse[dto.DebtTransactionResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /debts [post]
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

// HandleGetAll godoc
// @Summary      Get all debt transactions
// @Tags         debts
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.DebtTransactionResponse]
// @Failure      401  {object}  map[string]any
// @Router       /debts [get]
func (dh *DebtHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetTransactions(ctx.Request.Context(), profileID)
	})
}

// HandleGetTransactionSummary godoc
// @Summary      Get debt transaction summary
// @Tags         debts
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[map[string]dto.FriendBalance]
// @Failure      401  {object}  map[string]any
// @Router       /debts/summary [get]
func (dh *DebtHandler) HandleGetTransactionSummary() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetTransactionSummary", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetTransactionSummary(ctx.Request.Context(), profileID)
	})
}

// HandleGetRecent godoc
// @Summary      Get recent debt transactions
// @Tags         debts
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.DebtTransactionResponse]
// @Failure      401  {object}  map[string]any
// @Router       /debts/recent [get]
func (dh *DebtHandler) HandleGetRecent() gin.HandlerFunc {
	return server.Handler("DebtHandler.HandleGetRecent", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return dh.debtService.GetRecent(ctx.Request.Context(), profileID)
	})
}
