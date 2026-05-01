package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
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

// HandleCreateAnonymousFriendship godoc
// @Summary      Create an anonymous friendship
// @Tags         friendships
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.NewAnonymousFriendshipRequest true "New anonymous friendship payload"
// @Success      201  {object}  response.JSONResponse[dto.FriendshipResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /friendships [post]
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

// HandleGetAll godoc
// @Summary      Get all friendships
// @Tags         friendships
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.FriendshipResponse]
// @Failure      401  {object}  map[string]any
// @Router       /friendships [get]
func (fh *FriendshipHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("FriendshipHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return fh.friendshipService.GetAll(ctx.Request.Context(), profileID)
	})
}

// HandleGetDetails godoc
// @Summary      Get friendship details
// @Tags         friendships
// @Security     BearerAuth
// @Produce      json
// @Param        friendshipId path string true "Friendship ID"
// @Success      200  {object}  response.JSONResponse[dto.FriendDetailsResponse]
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /friendships/{friendshipId} [get]
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
