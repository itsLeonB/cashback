package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type PushSubscriptionHandler struct {
	pushNotificationSvc service.PushNotificationService
}

func NewPushSubscriptionHandler(pushSubscriptionService service.PushNotificationService) *PushSubscriptionHandler {
	return &PushSubscriptionHandler{pushSubscriptionService}
}

func (h *PushSubscriptionHandler) HandleSubscribe() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.PushSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID

		return nil, h.pushNotificationSvc.Subscribe(ctx, req)
	})
}

func (h *PushSubscriptionHandler) HandleUnsubscribe() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.PushUnsubscribeRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID

		return nil, h.pushNotificationSvc.Unsubscribe(ctx, req)
	})
}
