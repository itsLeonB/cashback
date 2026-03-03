package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
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

func (geh *OtherFeeHandler) HandleGetFeeCalculationMethods() gin.HandlerFunc {
	return server.Handler("OtherFeeHandler.HandleGetFeeCalculationMethods", http.StatusOK, func(ctx *gin.Context) (any, error) {
		return geh.otherFeeSvc.GetCalculationMethods(), nil
	})
}
