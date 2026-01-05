package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/ocr"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/ungerr"
)

type expenseBillServiceImpl struct {
	bucketName string
	taskQueue  meq.TaskQueue[message.ExpenseBillUploaded]
	billRepo   crud.Repository[expenses.ExpenseBill]
	transactor crud.Transactor
	imageSvc   storage.ImageService
	ocrSvc     ocr.OCRService
}

func NewExpenseBillService(
	bucketName string,
	taskQueue meq.TaskQueue[message.ExpenseBillUploaded],
	billRepo crud.Repository[expenses.ExpenseBill],
	transactor crud.Transactor,
	imageSvc storage.ImageService,
	ocrSvc ocr.OCRService,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		bucketName,
		taskQueue,
		billRepo,
		transactor,
		imageSvc,
		ocrSvc,
	}
}

func (ebs *expenseBillServiceImpl) Save(ctx context.Context, req *dto.NewExpenseBillRequest) (dto.ExpenseBillResponse, error) {
	var response dto.ExpenseBillResponse
	err := ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		fileID := ebs.objectKeyToFileID(util.GenerateObjectKey(req.Filename))
		newBill := expenses.ExpenseBill{
			GroupExpenseID: req.GroupExpenseID,
			ImageName:      fileID.ObjectKey,
			Status:         expenses.PendingBill,
		}

		savedBill, err := ebs.billRepo.Insert(ctx, newBill)
		if err != nil {
			return err
		}

		billUri, err := ebs.doUpload(ctx, req, fileID)
		if err != nil {
			return err
		}

		msg := message.ExpenseBillUploaded{
			ID:  savedBill.ID,
			URI: billUri,
		}

		if err = ebs.taskQueue.Enqueue(ctx, config.AppName, msg); err != nil {
			go ebs.rollbackUpload(ctx, fileID)
			return err
		}

		response = mapper.ExpenseBillToResponse(savedBill)

		return nil
	})

	return response, err
}

func (ebs *expenseBillServiceImpl) GetURL(ctx context.Context, billName string) (string, error) {
	return ebs.imageSvc.GetURL(ctx, ebs.objectKeyToFileID(billName))
}

func (ebs *expenseBillServiceImpl) ExtractBillText(ctx context.Context, msg message.ExpenseBillUploaded) (string, error) {
	var extractedText string
	err := ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[expenses.ExpenseBill]{}
		spec.Model.ID = msg.ID
		spec.ForUpdate = true
		bill, err := ebs.billRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if bill.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("expense bill with ID %s is not found", spec.Model.ID))
		}

		text, err := ebs.ocrSvc.ExtractFromURI(ctx, msg.URI)
		if err != nil {
			bill.Status = expenses.FailedExtracting
			_, statusErr := ebs.billRepo.Update(ctx, bill)
			if statusErr != nil {
				return errors.Join(err, statusErr)
			}
			// Don't return error here, just log the error
			logger.Errorf("failed to extract bill text: %v", err)
			return nil
		}

		extractedText = text
		bill.ExtractedText = text
		bill.Status = expenses.ExtractedBill
		_, err = ebs.billRepo.Update(ctx, bill)
		return err
	})
	return extractedText, err
}

func (ebs *expenseBillServiceImpl) Cleanup(ctx context.Context) error {
	spec := crud.Specification[expenses.ExpenseBill]{}
	bills, err := ebs.billRepo.FindAll(ctx, spec)
	if err != nil {
		return err
	}

	if len(bills) < 1 {
		logger.Info("no bills available")
		return nil
	}

	validObjectKeys := ezutil.MapSlice(bills, func(eb expenses.ExpenseBill) string { return eb.ImageName })

	logger.Infof("obtained object keys from DB:\n%s", strings.Join(validObjectKeys, "\n"))

	return ebs.imageSvc.DeleteAllInvalid(ctx, ebs.bucketName, validObjectKeys)
}

func (ebs *expenseBillServiceImpl) rollbackUpload(ctx context.Context, fileID storage.FileIdentifier) {
	if err := ebs.imageSvc.Delete(ctx, fileID); err != nil {
		logger.Errorf("error rolling back bill upload: %v", err)
	}
}

func (ebs *expenseBillServiceImpl) doUpload(
	ctx context.Context,
	req *dto.NewExpenseBillRequest,
	fileID storage.FileIdentifier,
) (string, error) {
	data, err := io.ReadAll(req.ImageReader)
	if err != nil {
		return "", ungerr.Wrap(err, "error reading image data")
	}

	return ebs.imageSvc.Upload(ctx, &storage.ImageUploadRequest{
		ImageData:      data,
		ContentType:    req.ContentType,
		FileSize:       req.FileSize,
		FileIdentifier: fileID,
	})
}

func (ebs *expenseBillServiceImpl) objectKeyToFileID(objectKey string) storage.FileIdentifier {
	return storage.FileIdentifier{
		BucketName: ebs.bucketName,
		ObjectKey:  objectKey,
	}
}
