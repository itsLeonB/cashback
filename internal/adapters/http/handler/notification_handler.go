package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type NotificationHandler struct {
	notificationService service.NotificationService
}

func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService}
}

func (nh *NotificationHandler) HandleGetUnread() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return nh.notificationService.GetUnread(ctx, profileID)
	})
}

func (nh *NotificationHandler) HandleMarkAsRead() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		notificationID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextNotificationID.String())
		if err != nil {
			return nil, err
		}

		return nil, nh.notificationService.MarkAsRead(ctx, profileID, notificationID)
	})
}

func (nh *NotificationHandler) HandleMarkAllAsRead() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return nil, nh.notificationService.MarkAllAsRead(ctx, profileID)
	})
}
