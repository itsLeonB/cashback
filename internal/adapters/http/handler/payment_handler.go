package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type PaymentHandler struct {
	svc service.PaymentService
}

func (ph *PaymentHandler) HandleWebhook() gin.HandlerFunc {
	return server.Handler("PaymentHandler.HandleWebhook", http.StatusOK, func(ctx *gin.Context) (any, error) {
		payload, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return nil, ungerr.Wrap(err, "error reading request body")
		}

		signature := ctx.GetHeader("Stripe-Signature")

		err = ph.svc.HandleWebhook(ctx.Request.Context(), payload, signature)
		if err != nil {
			return nil, err
		}

		return gin.H{"received": true}, nil
	})
}
