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

type groupExpenseHandler struct {
	groupExpenseService service.GroupExpenseService
}

func newGroupExpenseHandler(
	groupExpenseService service.GroupExpenseService,
) *groupExpenseHandler {
	return &groupExpenseHandler{
		groupExpenseService,
	}
}

func (geh *groupExpenseHandler) HandleCreateDraft() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request, err := ginkgo.BindRequest[dto.NewGroupExpenseRequest](ctx, binding.JSON)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		request.CreatorProfileID = userProfileID

		response, err := geh.groupExpenseService.CreateDraft(ctx, request)
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

func (geh *groupExpenseHandler) HandleGetAllCreated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		groupExpenses, err := geh.groupExpenseService.GetAllCreated(ctx, userProfileID)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(
			http.StatusOK,
			ginkgo.NewResponse(appconstant.MsgGetData).WithData(groupExpenses),
		)
	}
}

func (geh *groupExpenseHandler) HandleGetDetails() gin.HandlerFunc {
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

		response, err := geh.groupExpenseService.GetDetails(ctx, groupExpenseID, userProfileID)
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

func (geh *groupExpenseHandler) HandleConfirmDraft() gin.HandlerFunc {
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

		response, err := geh.groupExpenseService.ConfirmDraft(ctx, groupExpenseID, userProfileID)
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

func (geh *groupExpenseHandler) HandleCreateDraftV2() gin.HandlerFunc {
	return ginkgo.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		req, err := util.BindJSON[dto.NewDraftRequest](ctx)
		if err != nil {
			return nil, err
		}

		return geh.groupExpenseService.CreateDraftV2(ctx, userProfileID, req.Description)
	})
}

func (geh *groupExpenseHandler) HandleDelete() gin.HandlerFunc {
	return ginkgo.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.groupExpenseService.Delete(ctx, userProfileID, expenseID)
	})
}

func (geh *groupExpenseHandler) HandleSyncParticipants() gin.HandlerFunc {
	return ginkgo.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		req, err := ginkgo.BindJSON[dto.ExpenseParticipantsRequest](ctx)
		if err != nil {
			return nil, err
		}

		req.UserProfileID = userProfileID
		req.GroupExpenseID = expenseID

		return nil, geh.groupExpenseService.SyncParticipants(ctx, req)
	})
}
