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

type ExpenseBillHandler struct {
	expenseBillService service.ExpenseBillService
}

func NewExpenseBillHandler(expenseBillService service.ExpenseBillService) *ExpenseBillHandler {
	return &ExpenseBillHandler{expenseBillService}
}

// HandlePresignedSave godoc
// @Summary      Get a presigned URL to upload an expense bill
// @Tags         expense-bills
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        body body dto.PresignedExpenseBillRequest true "Presigned bill payload"
// @Success      201  {object}  response.JSONResponse[dto.PresignedExpenseBillResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/bills [post]
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

// HandleTriggerParsing godoc
// @Summary      Trigger parsing of an uploaded expense bill
// @Tags         expense-bills
// @Security     BearerAuth
// @Produce      json
// @Param        groupExpenseId path string true "Group expense ID"
// @Param        expenseBillId  path string true "Expense bill ID"
// @Success      200  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /group-expenses/{groupExpenseId}/bills/{expenseBillId} [put]
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
