package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

func getProfileID(ctx *gin.Context) (uuid.UUID, error) {
	return server.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextProfileID.String())
}
