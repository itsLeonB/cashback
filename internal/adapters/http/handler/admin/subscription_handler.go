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

type SubscriptionHandler struct {
	svc service.SubscriptionService
}

func (sh *SubscriptionHandler) HandleCreate() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.NewSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		return sh.svc.Create(ctx, req)
	})
}

func (sh *SubscriptionHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		subscriptions, err := sh.svc.GetList(ctx)
		if err != nil {
			return nil, err
		}

		ctx.Header("X-Total-Count", fmt.Sprint(len(subscriptions)))

		return subscriptions, nil
	})
}

func (sh *SubscriptionHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return sh.svc.GetOne(ctx, id)
	})
}

func (sh *SubscriptionHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.UpdateSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ID = id

		return sh.svc.Update(ctx, req)
	})
}

func (sh *SubscriptionHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return sh.svc.Delete(ctx, id)
	})
}
