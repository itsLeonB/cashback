package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
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

// HandleSubscribe godoc
// @Summary      Subscribe to push notifications
// @Tags         push
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.PushSubscriptionRequest true "Push subscription payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /push/subscribe [post]
func (h *PushSubscriptionHandler) HandleSubscribe() gin.HandlerFunc {
	return server.Handler("PushSubscriptionHandler.HandleSubscribe", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		sessionID, err := server.GetFromContext[uuid.UUID](ctx, appconstant.ContextSessionID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.PushSubscriptionRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID
		req.SessionID = sessionID

		return nil, h.pushNotificationSvc.Subscribe(ctx.Request.Context(), req)
	})
}

// HandleUnsubscribe godoc
// @Summary      Unsubscribe from push notifications
// @Tags         push
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.PushUnsubscribeRequest true "Push unsubscribe payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /push/unsubscribe [post]
func (h *PushSubscriptionHandler) HandleUnsubscribe() gin.HandlerFunc {
	return server.Handler("PushSubscriptionHandler.HandleUnsubscribe", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.PushUnsubscribeRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID

		return nil, h.pushNotificationSvc.Unsubscribe(ctx.Request.Context(), req)
	})
}
