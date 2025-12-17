package util

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/rotisserie/eris"
)

func GetProfileID(ctx context.Context) (uuid.UUID, error) {
	ctxProfileID := ctx.Value(appconstant.ContextProfileID.String())
	if ctxProfileID == nil {
		return uuid.Nil, eris.New("profileID not found in ctx")
	}

	switch id := ctxProfileID.(type) {
	case uuid.UUID:
		return id, nil
	case string:
		return ezutil.Parse[uuid.UUID](id)
	default:
		return uuid.Nil, eris.Errorf("unknown profileID format: %T", id)
	}
}

func BindJSON[T any](ctx *gin.Context) (T, error) {
	return ginkgo.BindRequest[T](ctx, binding.JSON)
}
