package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PlanHandler struct {
	svc service.PlanVersionService
}

// HandleGetActive godoc
// @Summary      Get active subscription plans
// @Tags         plans
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]monetization.PlanVersionResponse]
// @Router       /plans [get]
func (ph *PlanHandler) HandleGetActive() gin.HandlerFunc {
	return server.Handler("PlanHandler.HandleGetActive", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return ph.svc.GetActive(ctx.Request.Context())
	})
}
