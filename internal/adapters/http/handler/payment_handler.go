package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
