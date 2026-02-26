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

type PaymentHandler struct {
	svc service.PaymentService
}

func (ph *PaymentHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		payments, err := ph.svc.GetList(ctx)
		if err != nil {
			return nil, err
		}

		ctx.Header("X-Total-Count", fmt.Sprint(len(payments)))

		return payments, nil
	})
}

func (ph *PaymentHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPaymentID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.GetOne(ctx, id)
	})
}

func (ph *PaymentHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPaymentID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.UpdatePaymentRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ID = id

		return ph.svc.Update(ctx, req)
	})
}

func (ph *PaymentHandler) HandleDelete() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPaymentID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.Delete(ctx, id)
	})
}
