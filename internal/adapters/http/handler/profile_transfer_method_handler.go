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

type ProfileTransferMethodHandler struct {
	svc service.ProfileTransferMethodService
}

// HandleAdd godoc
// @Summary      Add a transfer method to profile
// @Tags         profile-transfer-methods
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.NewProfileTransferMethodRequest true "New profile transfer method payload"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /profile/transfer-methods [post]
func (ptmh *ProfileTransferMethodHandler) HandleAdd() gin.HandlerFunc {
	return server.Handler("ProfileTransferMethodHandler.HandleAdd", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.NewProfileTransferMethodRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID

		return nil, ptmh.svc.Add(ctx.Request.Context(), req)
	})
}

// HandleGetAllOwned godoc
// @Summary      Get all transfer methods owned by current profile
// @Tags         profile-transfer-methods
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.ProfileTransferMethodResponse]
// @Failure      401  {object}  map[string]any
// @Router       /profile/transfer-methods [get]
func (ptmh *ProfileTransferMethodHandler) HandleGetAllOwned() gin.HandlerFunc {
	return server.Handler("ProfileTransferMethodHandler.HandleGetAllOwned", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return ptmh.svc.GetAllByProfileID(ctx.Request.Context(), profileID)
	})
}

// HandleGetAllByFriendProfileID godoc
// @Summary      Get all transfer methods of a friend profile
// @Tags         profile-transfer-methods
// @Security     BearerAuth
// @Produce      json
// @Param        profileId path string true "Friend profile ID"
// @Success      200  {object}  response.JSONResponse[[]dto.ProfileTransferMethodResponse]
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /profiles/{profileId}/transfer-methods [get]
func (ptmh *ProfileTransferMethodHandler) HandleGetAllByFriendProfileID() gin.HandlerFunc {
	return server.Handler("ProfileTransferMethodHandler.HandleGetAllByFriendProfileID", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		friendProfileID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return ptmh.svc.GetAllByFriendProfileID(ctx.Request.Context(), userProfileID, friendProfileID)
	})
}
