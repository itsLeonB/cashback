package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PlanHandler struct {
	svc service.PlanService
}

func (ph *PlanHandler) HandleCreate() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.NewPlanRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ph.svc.Create(ctx, req)
	})
}

func (ph *PlanHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		return ph.svc.GetList(ctx)
	})
}

func (ph *PlanHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.GetOne(ctx, id)
	})
}

func (ph *PlanHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.UpdatePlanRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ID = id

		return ph.svc.Update(ctx, req)
	})
}

func (ph *PlanHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanID.String())
		if err != nil {
			return nil, err
		}

		return nil, ph.svc.Delete(ctx, id)
	})
}
