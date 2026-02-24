package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type SubscriptionHandler struct {
	svc service.SubscriptionService
}

func (sh *SubscriptionHandler) HandleCreatePurchase() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		planID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanID.String())
		if err != nil {
			return nil, err
		}

		planVersionID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextPlanVersionID.String())
		if err != nil {
			return nil, err
		}

		req := dto.PurchaseSubscriptionRequest{
			ProfileID:     profileID,
			PlanID:        planID,
			PlanVersionID: planVersionID,
		}

		return sh.svc.CreatePurchase(ctx, req)
	})
}
