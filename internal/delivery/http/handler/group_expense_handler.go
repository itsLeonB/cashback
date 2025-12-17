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

type GroupExpenseHandler struct {
	groupExpenseService service.GroupExpenseService
}

func NewGroupExpenseHandler(
	groupExpenseService service.GroupExpenseService,
) *GroupExpenseHandler {
	return &GroupExpenseHandler{
		groupExpenseService,
	}
}

func (geh *GroupExpenseHandler) HandleCreateDraft() gin.HandlerFunc {
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

func (geh *GroupExpenseHandler) HandleGetAllCreated() gin.HandlerFunc {
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

func (geh *GroupExpenseHandler) HandleGetDetails() gin.HandlerFunc {
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

func (geh *GroupExpenseHandler) HandleConfirmDraft() gin.HandlerFunc {
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

func (geh *GroupExpenseHandler) HandleCreateDraftV2() gin.HandlerFunc {
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
