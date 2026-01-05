package util

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ungerr"
)

func GetProfileID(ctx context.Context) (uuid.UUID, error) {
	ctxProfileID := ctx.Value(appconstant.ContextProfileID.String())
	if ctxProfileID == nil {
		return uuid.Nil, ungerr.Unknown("profileID not found in ctx")
	}

	switch id := ctxProfileID.(type) {
	case uuid.UUID:
		return id, nil
	case string:
		return ezutil.Parse[uuid.UUID](id)
	default:
		return uuid.Nil, ungerr.Unknownf("unknown profileID format: %T", id)
	}
}
