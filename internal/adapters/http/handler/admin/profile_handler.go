package admin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type ProfileHandler struct {
	svc service.ProfileService
}

func (ph *ProfileHandler) HandleGetList() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		profiles, err := ph.svc.GetAllReal(ctx)
		if err != nil {
			return nil, err
		}

		ctx.Header("X-Total-Count", fmt.Sprint(len(profiles)))

		return profiles, nil
	})
}

func (ph *ProfileHandler) HandleGetOne() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		id, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return ph.svc.GetByID(ctx, id)
	})
}
