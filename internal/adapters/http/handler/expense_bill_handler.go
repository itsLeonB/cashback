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

type ExpenseBillHandler struct {
	expenseBillService service.ExpenseBillService
}

func NewExpenseBillHandler(expenseBillService service.ExpenseBillService) *ExpenseBillHandler {
	return &ExpenseBillHandler{expenseBillService}
}

func (geh *ExpenseBillHandler) HandlePresignedSave() gin.HandlerFunc {
	return server.Handler("ExpenseBillHandler.HandlePresignedSave", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		req, err := server.BindJSON[dto.PresignedExpenseBillRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.ProfileID = profileID
		req.GroupExpenseID = expenseID

		return geh.expenseBillService.SavePresigned(ctx.Request.Context(), req)
	})
}

func (geh *ExpenseBillHandler) HandleTriggerParsing() gin.HandlerFunc {
	return server.Handler("ExpenseBillHandler.HandleTriggerParsing", http.StatusOK, func(ctx *gin.Context) (any, error) {
		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		billID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseBillID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.expenseBillService.TriggerParsing(ctx.Request.Context(), expenseID, billID)
	})
}
