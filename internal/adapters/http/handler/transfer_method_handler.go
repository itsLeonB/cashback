package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type TransferMethodHandler struct {
	transferMethodService service.TransferMethodService
}

func NewTransferMethodHandler(transferMethodService service.TransferMethodService) *TransferMethodHandler {
	return &TransferMethodHandler{transferMethodService}
}

// HandleGetAll godoc
// @Summary      Get all transfer methods
// @Tags         transfer-methods
// @Security     BearerAuth
// @Produce      json
// @Param        status query string false "Filter by status"
// @Success      200  {object}  response.JSONResponse[[]dto.TransferMethodResponse]
// @Failure      401  {object}  map[string]any
// @Router       /transfer-methods [get]
func (tmh *TransferMethodHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("TransferMethodHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return tmh.transferMethodService.GetAll(ctx.Request.Context(), debts.ParentFilter(ctx.Query("status")), profileID)
	})
}
