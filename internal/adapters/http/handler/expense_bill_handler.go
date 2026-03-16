package handler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Flagsmith/flagsmith-go-client"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type ExpenseBillHandler struct {
	expenseBillService service.ExpenseBillService
	flagClient         *flagsmith.Client
}

func NewExpenseBillHandler(expenseBillService service.ExpenseBillService) *ExpenseBillHandler {
	client := flagsmith.NewClient(config.Global.ClientKey, flagsmith.DefaultConfig())
	return &ExpenseBillHandler{expenseBillService, client}
}

func (geh *ExpenseBillHandler) HandleSave() gin.HandlerFunc {
	usePresigned, err := geh.flagClient.FeatureEnabled("use_presigned_bill_upload")
	if err != nil {
		logger.Error(ungerr.Wrap(err, "error fetching feature flag"))
	}
	if usePresigned {
		return geh.HandlePresignedSave()
	}
	return server.Handler("ExpenseBillHandler.HandleSave", http.StatusCreated, func(ctx *gin.Context) (any, error) {
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

		if fileHeader.Size > storage.MaxFileSize {
			return nil, ungerr.UnprocessableEntityError("file too large")
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

		return nil, geh.expenseBillService.Save(ctx.Request.Context(), request)
	})
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
