package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type PublicHandler struct {
	friendDetailsSvc service.FriendDetailsService
}

func NewPublicHandler(friendDetailsSvc service.FriendDetailsService) *PublicHandler {
	return &PublicHandler{friendDetailsSvc}
}

func (ph *PublicHandler) HandleGetPublicProfile() gin.HandlerFunc {
	return server.Handler("PublicHandler.HandleGetPublicProfile", http.StatusOK, func(ctx *gin.Context) (any, error) {
		slug := ctx.Param("slug")
		if slug == "" {
			return nil, ungerr.BadRequestError("slug is required")
		}
		return ph.friendDetailsSvc.GetDetailsBySlug(ctx.Request.Context(), slug)
	})
}
