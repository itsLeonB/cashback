package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
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

func (ph *ProfileHandler) HandleProfile() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := server.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return ph.profileService.GetByID(ctx, profileID)
	})
}

func (ph *ProfileHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := server.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.UpdateProfileRequest](ctx)
		if err != nil {
			return nil, err
		}

		return ph.profileService.Update(ctx, profileID, request.Name)
	})
}

func (ph *ProfileHandler) HandleSearch() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}
		request, err := server.BindRequest[dto.SearchRequest](ctx, binding.Query)
		if err != nil {
			return nil, err
		}

		return ph.profileService.Search(ctx, profileID, request.Query)
	})
}

func (ph *ProfileHandler) HandleAssociate() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.AssociateProfileRequest](ctx)
		if err != nil {
			return nil, err
		}

		return nil, ph.profileService.Associate(ctx, profileID, request.RealProfileID, request.AnonProfileID)
	})
}
