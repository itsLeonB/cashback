package handler

import (
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

func (ph *PaymentHandler) HandleNotification() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.MidtransNotificationPayload](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ph.svc.HandleNotification(ctx, req)
	})
}

func (ph *PaymentHandler) HandleMakePayment() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		subscriptionID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.MakePayment(ctx, subscriptionID)
	})
}
