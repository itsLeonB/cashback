package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
)

type OtherFeeHandler struct {
	otherFeeSvc service.OtherFeeService
}

func NewOtherFeeHandler(
	otherFeeSvc service.OtherFeeService,
) *OtherFeeHandler {
	return &OtherFeeHandler{
		otherFeeSvc,
	}
}

// HandleAdd godoc
// @Summary      Add a fee to a group expense
// @Tags         other-fees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        body body dto.NewOtherFeeRequest true "New fee payload"
// @Success      201  {object}  response.JSONResponse[dto.OtherFeeResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/fees [post]
func (geh *OtherFeeHandler) HandleAdd() gin.HandlerFunc {
	return server.Handler("OtherFeeHandler.HandleAdd", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.NewOtherFeeRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID

		return geh.otherFeeSvc.Add(ctx.Request.Context(), request)
	})
}

// HandleUpdate godoc
// @Summary      Update a fee on a group expense
// @Tags         other-fees
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        otherFeeId     path string true "Fee ID"
// @Param        body body dto.UpdateOtherFeeRequest true "Update fee payload"
// @Success      200  {object}  response.JSONResponse[dto.OtherFeeResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/fees/{otherFeeId} [put]
func (geh *OtherFeeHandler) HandleUpdate() gin.HandlerFunc {
	return server.Handler("OtherFeeHandler.HandleUpdate", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		otherFeeID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextOtherFeeID.String())
		if err != nil {
			return nil, err
		}

		request, err := server.BindJSON[dto.UpdateOtherFeeRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID
		request.ID = otherFeeID

		return geh.otherFeeSvc.Update(ctx.Request.Context(), request)
	})
}

// HandleRemove godoc
// @Summary      Remove a fee from a group expense
// @Tags         other-fees
// @Security     BearerAuth
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        otherFeeId     path string true "Fee ID"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/fees/{otherFeeId} [delete]
func (geh *OtherFeeHandler) HandleRemove() gin.HandlerFunc {
	return server.Handler("OtherFeeHandler.HandleRemove", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		feeID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextOtherFeeID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.otherFeeSvc.Remove(ctx.Request.Context(), groupExpenseID, feeID, userProfileID)
	})
}

// HandleGetFeeCalculationMethods godoc
// @Summary      Get available fee calculation methods
// @Tags         other-fees
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.JSONResponse[[]dto.FeeCalculationMethodInfo]
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/fee-calculation-methods [get]
func (geh *OtherFeeHandler) HandleGetFeeCalculationMethods() gin.HandlerFunc {
	return server.Handler("OtherFeeHandler.HandleGetFeeCalculationMethods", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return geh.otherFeeSvc.GetCalculationMethods(), nil
	})
}
