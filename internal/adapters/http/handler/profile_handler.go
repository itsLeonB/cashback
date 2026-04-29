package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type ProfileHandler struct {
	profileService service.ProfileService
}

func NewProfileHandler(
	profileService service.ProfileService,
) *ProfileHandler {
	return &ProfileHandler{
		profileService,
	}
}

// HandleProfile godoc
// @Summary      Get current user's profile
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[dto.ProfileResponse]
// @Failure      401  {object}  map[string]any
// @Router       /profile [get]
func (ph *ProfileHandler) HandleProfile() gin.HandlerFunc {
	return server.Handler("ProfileHandler.HandleProfile", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := server.GetFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return ph.profileService.GetByID(ctx.Request.Context(), profileID)
	})
}

// HandleUpdate godoc
// @Summary      Update current user's profile
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.UpdateProfileRequest true "Update profile payload"
// @Success      200  {object}  response.JSONResponse[dto.ProfileResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /profile [patch]
func (ph *ProfileHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler("ProfileHandler.HandleUpdate", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := server.GetFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.UpdateProfileRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.ID = profileID

		return ph.profileService.Update(ctx.Request.Context(), request)
	})
}

// HandleSearch godoc
// @Summary      Search profiles
// @Tags         profile
// @Security     BearerAuth
// @Produce      json
// @Param        q query string false "Search query"
// @Success      200  {object}  response.JSONResponse[[]dto.ProfileResponse]
// @Failure      401  {object}  map[string]any
// @Router       /profiles [get]
func (ph *ProfileHandler) HandleSearch() gin.HandlerFunc {
	return server.Handler("ProfileHandler.HandleSearch", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}
		request, err := server.BindRequest[dto.SearchRequest](ctx, binding.Query)
		if err != nil {
			return nil, err
		}

		return ph.profileService.Search(ctx.Request.Context(), profileID, request.Query)
	})
}

// HandleAssociate godoc
// @Summary      Associate anonymous profile with real profile
// @Tags         profile
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.AssociateProfileRequest true "Associate payload"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /profile/associate [post]
func (ph *ProfileHandler) HandleAssociate() gin.HandlerFunc {
	return server.Handler("ProfileHandler.HandleAssociate", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.AssociateProfileRequest](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ph.profileService.Associate(ctx.Request.Context(), profileID, request.RealProfileID, request.AnonProfileID)
	})
}
