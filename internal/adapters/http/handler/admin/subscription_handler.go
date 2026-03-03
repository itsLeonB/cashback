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
	return server.Handler("SubscriptionHandler.HandleCreate", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.NewSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		return sh.svc.Create(ctx.Request.Context(), req)
	})
}

func (sh *SubscriptionHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleGetList", http.StatusOK, func(ctx *gin.Context) (any, error) {
		subscriptions, err := sh.svc.GetList(ctx.Request.Context())
		if err != nil {
			return nil, err
		}

		ctx.Header("X-Total-Count", fmt.Sprint(len(subscriptions)))

		return subscriptions, nil
	})
}

func (sh *SubscriptionHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleGetOne", http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return sh.svc.GetOne(ctx.Request.Context(), id)
	})
}

func (sh *SubscriptionHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleUpdate", http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.UpdateSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ID = id

		return sh.svc.Update(ctx.Request.Context(), req)
	})
}

func (sh *SubscriptionHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleDelete", http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return sh.svc.Delete(ctx.Request.Context(), id)
	})
}
