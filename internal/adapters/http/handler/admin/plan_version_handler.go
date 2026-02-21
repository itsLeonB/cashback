package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PlanVersionHandler struct {
	svc service.PlanVersionService
}

func (pvh *PlanVersionHandler) HandleCreate() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.NewPlanVersionRequest](ctx)
		if err != nil {
			return nil, err
		}

		return pvh.svc.Create(ctx, req)
	})
}

func (pvh *PlanVersionHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		planVersions, err := pvh.svc.GetList(ctx)
		if err != nil {
			return nil, err
		}

		ctx.Header("X-Total-Count", fmt.Sprint(len(planVersions)))

		return planVersions, nil
	})
}

func (pvh *PlanVersionHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanVersionID.String())
		if err != nil {
			return nil, err
		}

		return pvh.svc.GetOne(ctx, id)
	})
}

func (pvh *PlanVersionHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanVersionID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.UpdatePlanVersionRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ID = id

		return pvh.svc.Update(ctx, req)
	})
}

func (pvh *PlanVersionHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanVersionID.String())
		if err != nil {
			return nil, err
		}

		return pvh.svc.Delete(ctx, id)
	})
}
