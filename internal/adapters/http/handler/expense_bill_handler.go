package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type ExpenseBillHandler struct {
	expenseBillService service.ExpenseBillService
}

func NewExpenseBillHandler(expenseBillService service.ExpenseBillService) *ExpenseBillHandler {
	return &ExpenseBillHandler{expenseBillService}
}

func (geh *ExpenseBillHandler) HandleSave() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		profileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		fileHeader, err := ctx.FormFile("bill")
		if err != nil {
			return nil, ungerr.Wrap(err, appconstant.ErrProcessFile)
		}

		file, err := fileHeader.Open()
		if err != nil {
			return nil, ungerr.Wrap(err, appconstant.ErrProcessFile)
		}
		defer func() {
			if e := file.Close(); e != nil {
				logger.Errorf("error closing file reader: %v", e)
			}
		}()

		data, err := io.ReadAll(file)
		if err != nil {
			return nil, ungerr.Wrap(err, appconstant.ErrProcessFile)
		}

		if err := util.ValidateImage(bytes.NewReader(data), int64(len(data))); err != nil {
			return nil, err
		}

		request := &dto.NewExpenseBillRequest{
			ImageData:      data,
			ProfileID:      profileID,
			GroupExpenseID: expenseID,
			ContentType:    fileHeader.Header.Get("Content-Type"),
			Filename:       fileHeader.Filename,
			FileSize:       fileHeader.Size,
		}

		return nil, geh.expenseBillService.Save(ctx, request)
	})
}

func (geh *ExpenseBillHandler) HandleTriggerParsing() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		expenseID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextGroupExpenseID.String())
		if err != nil {
			return nil, err
		}

		billID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextExpenseBillID.String())
		if err != nil {
			return nil, err
		}

		return nil, geh.expenseBillService.TriggerParsing(ctx, expenseID, billID)
	})
}
