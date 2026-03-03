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

type FriendshipHandler struct {
	friendshipService service.FriendshipService
	friendDetailsSvc  service.FriendDetailsService
}

func NewFriendshipHandler(
	friendshipService service.FriendshipService,
	friendDetailsSvc service.FriendDetailsService,
) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService,
		friendDetailsSvc,
	}
}

func (fh *FriendshipHandler) HandleCreateAnonymousFriendship() gin.HandlerFunc {
	return server.Handler("FriendshipHandler.HandleCreateAnonymousFriendship", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.NewAnonymousFriendshipRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.ProfileID = profileID

		return fh.friendshipService.CreateAnonymous(ctx.Request.Context(), request)
	})
}

func (fh *FriendshipHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("FriendshipHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return fh.friendshipService.GetAll(ctx.Request.Context(), profileID)
	})
}

func (fh *FriendshipHandler) HandleGetDetails() gin.HandlerFunc {
	return server.Handler("FriendshipHandler.HandleGetDetails", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		friendshipID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextFriendshipID.String())
		if err != nil {
			return nil, err
		}

		return fh.friendDetailsSvc.GetDetails(ctx.Request.Context(), profileID, friendshipID)
	})
}
