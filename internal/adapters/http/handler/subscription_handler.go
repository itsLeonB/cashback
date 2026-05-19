package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	dto "github.com/itsLeonB/cashback/internal/domain/dto/monetization"
	service "github.com/itsLeonB/cashback/internal/domain/service/monetization"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type SubscriptionHandler struct {
	svc        service.SubscriptionService
	paymentSvc service.PaymentService
}

// HandleCreatePurchase godoc
// @Summary      Create a subscription purchase
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Param        planId        path string true "Plan ID"
// @Param        planVersionId path string true "Plan version ID"
// @Success      201  {object}  response.JSONResponse[monetization.PaymentResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /plans/{planId}/versions/{planVersionId}/subscriptions [post]
func (sh *SubscriptionHandler) HandleCreatePurchase() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleCreatePurchase", http.StatusCreated, func(ctx *gin.Context) (any, error) {
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

		return sh.paymentSvc.NewPurchase(ctx.Request.Context(), req)
	})
}

// HandlePortalSession godoc
// @Summary      Create a Stripe billing portal session
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[monetization.PortalSessionResponse]
// @Failure      401  {object}  map[string]any
// @Router       /subscriptions/portal [post]
func (sh *SubscriptionHandler) HandlePortalSession() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandlePortalSession", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return sh.paymentSvc.CreatePortalSession(ctx.Request.Context(), profileID)
	})
}

// HandleGetSubscribedDetails godoc
// @Summary      Get current subscription details
// @Tags         subscriptions
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[monetization.SubscriptionResponse]
// @Failure      401  {object}  map[string]any
// @Router       /profile/subscription [get]
func (sh *SubscriptionHandler) HandleGetSubscribedDetails() gin.HandlerFunc {
	return server.Handler("SubscriptionHandler.HandleGetSubscribedDetails", http.StatusOK, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		return sh.svc.GetSubscribedDetails(ctx.Request.Context(), profileID)
	})
}
