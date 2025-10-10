package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/service"
	"github.com/itsLeonB/orcashtrator/internal/util"
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
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		profileID, err := ginkgo.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return 0, "", nil, err
		}

		response, err := ph.profileService.GetByID(ctx, profileID)
		if err != nil {
			return 0, "", nil, err
		}

		return http.StatusOK, appconstant.MsgGetData, response, nil
	})
}

func (ph *ProfileHandler) HandleUpdate() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		profileID, err := ginkgo.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return 0, "", nil, err
		}

		request, err := ginkgo.BindRequest[dto.UpdateProfileRequest](ctx, binding.JSON)
		if err != nil {
			return 0, "", nil, err
		}

		response, err := ph.profileService.Update(ctx, profileID, request.Name)
		if err != nil {
			return 0, "", nil, err
		}

		return http.StatusOK, appconstant.MsgUpdateData, response, nil
	})
}

func (ph *ProfileHandler) HandleSearch() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		profileID, err := util.GetProfileID(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		request, err := ginkgo.BindRequest[dto.SearchRequest](ctx, binding.Query)
		if err != nil {
			return 0, "", nil, err
		}

		response, err := ph.profileService.Search(ctx, profileID, request.Query)
		if err != nil {
			return 0, "", nil, err
		}

		return http.StatusOK, appconstant.MsgGetData, response, nil
	})
}
