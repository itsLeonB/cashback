package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

func getProfileID(ctx *gin.Context) (uuid.UUID, error) {
	return server.GetFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
}
