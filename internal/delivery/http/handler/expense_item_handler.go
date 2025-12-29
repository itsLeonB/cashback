package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/service"
	"github.com/itsLeonB/orcashtrator/internal/util"
)

type ExpenseItemHandler struct {
	expenseItemSvc service.ExpenseItemService
}

func NewExpenseItemHandler(
	expenseItemSvc service.ExpenseItemService,
) *ExpenseItemHandler {
	return &ExpenseItemHandler{
		expenseItemSvc,
	}
}

func (geh *ExpenseItemHandler) HandleAdd() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		groupExpenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request, err := ginkgo.BindRequest[dto.NewExpenseItemRequest](ctx, binding.JSON)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID

		response, err := geh.expenseItemSvc.Add(ctx, request)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(
			http.StatusCreated,
			ginkgo.NewResponse(appconstant.MsgInsertData).WithData(response),
		)
	}
}

func (geh *ExpenseItemHandler) HandleGetDetails() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		groupExpenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		expenseItemID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		response, err := geh.expenseItemSvc.GetDetails(ctx, groupExpenseID, expenseItemID, userProfileID)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(
			http.StatusOK,
			ginkgo.NewResponse(appconstant.MsgGetData).WithData(response),
		)
	}
}

func (geh *ExpenseItemHandler) HandleUpdate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		groupExpenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		expenseItemID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request, err := ginkgo.BindRequest[dto.UpdateExpenseItemRequest](ctx, binding.JSON)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request.UserProfileID = userProfileID
		request.GroupExpenseID = groupExpenseID
		request.ID = expenseItemID

		response, err := geh.expenseItemSvc.Update(ctx, request)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(
			http.StatusOK,
			ginkgo.NewResponse(appconstant.MsgUpdateData).WithData(response),
		)
	}
}

func (geh *ExpenseItemHandler) HandleRemove() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		groupExpenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		expenseItemID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		if err = geh.expenseItemSvc.Remove(ctx, groupExpenseID, expenseItemID, userProfileID); err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func (geh *ExpenseItemHandler) HandleSyncParticipants() gin.HandlerFunc {
	return ginkgo.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		groupExpenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			_ = ctx.Error(err)
			return nil, err
		}

		expenseItemID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseItemID.String())
		if err != nil {
			return nil, err
		}

		request, err := ginkgo.BindJSON[dto.SyncItemParticipantsRequest](ctx)
		if err != nil {
			return nil, err
		}

		request.ProfileID = userProfileID
		request.ID = expenseItemID
		request.GroupExpenseID = groupExpenseID

		return nil, geh.expenseItemSvc.SyncParticipants(ctx, request)
	})
}
