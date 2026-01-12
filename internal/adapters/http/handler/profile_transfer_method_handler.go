package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type ProfileTransferMethodHandler struct {
	svc service.ProfileTransferMethodService
}

func (ptmh *ProfileTransferMethodHandler) HandleAdd() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.NewProfileTransferMethodRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID

		return nil, ptmh.svc.Add(ctx, req)
	})
}

func (ptmh *ProfileTransferMethodHandler) HandleGetAllOwned() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return ptmh.svc.GetAllByProfileID(ctx, profileID)
	})
}
