package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/cashback/internal/domain/service/monetization/payment"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PaymentHandler struct {
	svc service.PaymentService
}

// HandleMidtransNotification godoc
// @Summary      Handle Midtrans payment notification
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        body body dto.MidtransNotificationPayload true "Midtrans notification payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Router       /payments/midtrans/notifications [post]
func (ph *PaymentHandler) HandleMidtransNotification() gin.HandlerFunc {
	return server.Handler("PaymentHandler.HandleMidtransNotification", http.StatusOK, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.MidtransNotificationPayload](ctx)
		if err != nil {
			return nil, err
		}

		payload := payment.NotificationPayload{
			OrderID:   req.OrderID,
			Signature: req.SignatureKey,
			Extra: map[string]string{
				"status_code":    req.StatusCode,
				"gross_amount":   req.GrossAmount,
				"status_message": req.StatusMessage,
			},
		}

		return nil, ph.svc.HandleNotification(ctx.Request.Context(), payload)
	})
}

// HandleIPaymuNotification godoc
// @Summary      Handle iPaymu payment notification
// @Tags         payments
// @Accept       json
// @Produce      json
// @Param        body body dto.IPaymuNotificationPayload true "iPaymu notification payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Router       /payments/ipaymu/notifications [post]
func (ph *PaymentHandler) HandleIPaymuNotification() gin.HandlerFunc {
	return server.Handler("PaymentHandler.HandleIPaymuNotification", http.StatusOK, func(ctx *gin.Context) (any, error) {
		req, err := server.BindJSON[dto.IPaymuNotificationPayload](ctx)
		if err != nil {
			return nil, err
		}

		payload := payment.NotificationPayload{
			OrderID:   req.ReferenceID,
			RawStatus: req.Status,
			Signature: ctx.Query("secret"),
			Extra: map[string]string{
				"trx_id":      req.TrxID,
				"status_code": req.StatusCode,
			},
		}

		return nil, ph.svc.HandleNotification(ctx.Request.Context(), payload)
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
