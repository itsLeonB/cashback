package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService}
}

// HandleGetUnread godoc
// @Summary      Get unread notifications
// @Tags         notifications
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.NotificationResponse]
// @Failure      401  {object}  map[string]any
// @Router       /notifications [get]
func (nh *NotificationHandler) HandleGetUnread() gin.HandlerFunc {
	return server.Handler("NotificationHandler.HandleGetUnread", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return nh.notificationService.GetUnread(ctx.Request.Context(), profileID)
	})
}

// HandleMarkAsRead godoc
// @Summary      Mark a notification as read
// @Tags         notifications
// @Security     BearerAuth
// @Param        notificationId path string true "Notification ID"
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /notifications/{notificationId} [patch]
func (nh *NotificationHandler) HandleMarkAsRead() gin.HandlerFunc {
	return server.Handler("NotificationHandler.HandleMarkAsRead", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		notificationID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextNotificationID.String())
		if err != nil {
			return nil, err
		}

		return nil, nh.notificationService.MarkAsRead(ctx.Request.Context(), profileID, notificationID)
	})
}

// HandleMarkAllAsRead godoc
// @Summary      Mark all notifications as read
// @Tags         notifications
// @Security     BearerAuth
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /notifications [patch]
func (nh *NotificationHandler) HandleMarkAllAsRead() gin.HandlerFunc {
	return server.Handler("NotificationHandler.HandleMarkAllAsRead", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return nil, nh.notificationService.MarkAllAsRead(ctx.Request.Context(), profileID)
	})
}
