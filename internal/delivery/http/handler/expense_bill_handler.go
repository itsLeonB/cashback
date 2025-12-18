package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/service"
	"github.com/itsLeonB/orcashtrator/internal/util"
	"github.com/rotisserie/eris"
)

type ExpenseBillHandler struct {
	logger             ezutil.Logger
	expenseBillService service.ExpenseBillService
}

func NewExpenseBillHandler(
	logger ezutil.Logger,
	expenseBillService service.ExpenseBillService,
) *ExpenseBillHandler {
	return &ExpenseBillHandler{
		logger,
		expenseBillService,
	}
}

func (geh *ExpenseBillHandler) HandleSave() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		payerProfileID := uuid.Nil
		if payerProfileIDStr := ctx.PostForm("payerProfileId"); payerProfileIDStr != "" {
			payerProfileID, err = ezutil.Parse[uuid.UUID](payerProfileIDStr)
			if err != nil {
				_ = ctx.Error(err)
				return
			}
		}

		fileHeader, err := ctx.FormFile("bill")
		if err != nil {
			_ = ctx.Error(eris.Wrap(err, appconstant.ErrProcessFile))
			return
		}

		file, err := fileHeader.Open()
		if err != nil {
			_ = ctx.Error(eris.Wrap(err, appconstant.ErrProcessFile))
			return
		}
		defer func() {
			if e := file.Close(); e != nil {
				geh.logger.Errorf("error closing file reader: %v", e)
			}
		}()

		request := dto.NewExpenseBillRequest{
			CreatorProfileID: userProfileID,
			PayerProfileID:   payerProfileID,
			ImageReader:      file,
			ContentType:      fileHeader.Header.Get("Content-Type"),
			Filename:         fileHeader.Filename,
			FileSize:         fileHeader.Size,
		}

		response, err := geh.expenseBillService.Save(ctx, &request)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(
			http.StatusCreated,
			ginkgo.NewResponse(appconstant.MsgBillUploaded).WithData(response),
		)
	}
}

func (geh *ExpenseBillHandler) HandleGetAllCreated() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		response, err := geh.expenseBillService.GetAllCreated(ctx, userProfileID)
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

func (geh *ExpenseBillHandler) HandleGet() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		id, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseBillID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		response, err := geh.expenseBillService.Get(ctx, userProfileID, id)
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

func (geh *ExpenseBillHandler) HandleDelete() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		id, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseBillID.String())
		if err != nil {
			_ = ctx.Error(err)
			return
		}

		if err = geh.expenseBillService.Delete(ctx, userProfileID, id); err != nil {
			_ = ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusNoContent, nil)
	}
}

func (geh *ExpenseBillHandler) HandleSaveV2() gin.HandlerFunc {
	return ginkgo.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := ginkgo.GetAndParseFromContext[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		fileHeader, err := ctx.FormFile("bill")
		if err != nil {
			return nil, eris.Wrap(err, appconstant.ErrProcessFile)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, eris.Wrap(err, appconstant.ErrProcessFile)
		}
		defer func() {
			if e := file.Close(); e != nil {
				geh.logger.Errorf("error closing file reader: %v", e)
			}
		}()

		request := dto.NewExpenseBillRequest{
			CreatorProfileID: userProfileID,
			GroupExpenseID:   expenseID,
			ImageReader:      file,
			ContentType:      fileHeader.Header.Get("Content-Type"),
			Filename:         fileHeader.Filename,
			FileSize:         fileHeader.Size,
		}

		response, err := geh.expenseBillService.SaveV2(ctx, &request)
		if err != nil {
			return nil, err
		}

		return response, nil
	})
}
