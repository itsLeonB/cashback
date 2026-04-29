package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PaymentHandler struct {
	svc service.PaymentService
}

// HandleNotification godoc
// @Summary      Handle Midtrans payment notification
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        body body dto.MidtransNotificationPayload true "Midtrans notification payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Router       /payments/midtrans/notifications [post]
func (ph *PaymentHandler) HandleNotification() gin.HandlerFunc {
	return server.Handler("PaymentHandler.HandleNotification", http.StatusOK, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.MidtransNotificationPayload](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ph.svc.HandleNotification(ctx.Request.Context(), req)
	})
}

// HandleMakePayment godoc
// @Summary      Make a payment for a subscription
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        subscriptionId path string true "Subscription ID"
// @Success      201  {object}  response.JSONResponse[monetization.PaymentResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /subscriptions/{subscriptionId} [post]
func (ph *PaymentHandler) HandleMakePayment() gin.HandlerFunc {
	return server.Handler("PaymentHandler.HandleMakePayment", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		subscriptionID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextSubscriptionID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.MakePayment(ctx.Request.Context(), subscriptionID)
	})
}
