package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PlanHandler struct {
	svc service.PlanVersionService
}

func (ph *PlanHandler) HandleGetActive() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return ph.svc.GetActive(ctx)
	})
}
