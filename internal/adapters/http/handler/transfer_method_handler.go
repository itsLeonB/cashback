package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type TransferMethodHandler struct {
	transferMethodService service.TransferMethodService
}

func NewTransferMethodHandler(transferMethodService service.TransferMethodService) *TransferMethodHandler {
	return &TransferMethodHandler{transferMethodService}
}

func (tmh *TransferMethodHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return tmh.transferMethodService.GetAll(ctx)
	})
}
