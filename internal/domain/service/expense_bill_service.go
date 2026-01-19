package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/ocr"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type expenseBillServiceImpl struct {
	taskQueue  queue.TaskQueue
	billRepo   crud.Repository[expenses.ExpenseBill]
	transactor crud.Transactor
	imageSvc   storage.ImageService
	ocrSvc     ocr.OCRService
	expenseSvc GroupExpenseService
}

func NewExpenseBillService(
	taskQueue queue.TaskQueue,
	billRepo crud.Repository[expenses.ExpenseBill],
	transactor crud.Transactor,
	imageSvc storage.ImageService,
	ocrSvc ocr.OCRService,
	expenseSvc GroupExpenseService,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		taskQueue,
		billRepo,
		transactor,
		imageSvc,
		ocrSvc,
		expenseSvc,
	}
}

func (ebs *expenseBillServiceImpl) Save(ctx context.Context, req *dto.NewExpenseBillRequest) error {
	return ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := ebs.ensureSingleBill(ctx, req.ProfileID, req.GroupExpenseID); err != nil {
			return err
		}

		fileID := ObjectKeyToFileID(util.GenerateObjectKey(req.Filename))
		newBill := expenses.ExpenseBill{
			GroupExpenseID: req.GroupExpenseID,
			ImageName:      fileID.ObjectKey,
			Status:         expenses.PendingBill,
		}

		savedBill, err := ebs.billRepo.Insert(ctx, newBill)
		if err != nil {
			return err
		}

		if _, err = ebs.doUpload(ctx, req, fileID); err != nil {
			return err
		}

		if err = ebs.taskQueue.Enqueue(ctx, message.ExpenseBillUploaded{ID: savedBill.ID}); err != nil {
			go ebs.rollbackUpload(ctx, fileID)
			return err
		}

		return nil
	})
}

func (ebs *expenseBillServiceImpl) ensureSingleBill(ctx context.Context, profileID, expenseID uuid.UUID) error {
	expense, err := ebs.expenseSvc.GetUnconfirmedGroupExpenseForUpdate(ctx, profileID, expenseID)
	if err != nil {
		return err
	}
	if expense.Bill.IsZero() {
		return nil
	}

	if expense.Bill.Status != expenses.NotDetectedBill {
		return ungerr.UnprocessableEntityError("cannot upload another bill")
	}

	return ebs.billRepo.Delete(ctx, expense.Bill)
}

func (ebs *expenseBillServiceImpl) ExtractBillText(ctx context.Context, msg message.ExpenseBillUploaded) error {
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

		uri := ebs.imageSvc.GetURI(ObjectKeyToFileID(bill.ImageName))
		text, err := ebs.ocrSvc.ExtractFromURI(ctx, uri)
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

		bill.ExtractedText = text
		bill.Status = expenses.ExtractedBill
		_, err = ebs.billRepo.Update(ctx, bill)
		return err
	})
	if err != nil {
		return err
	}

	return ebs.taskQueue.Enqueue(ctx, message.ExpenseBillTextExtracted(msg))
}

func (ebs *expenseBillServiceImpl) TriggerParsing(ctx context.Context, expenseID, billID uuid.UUID) error {
	return ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[expenses.ExpenseBill]{}
		spec.Model.ID = billID
		spec.Model.GroupExpenseID = expenseID
		spec.ForUpdate = true
		bill, err := ebs.billRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if bill.IsZero() {
			return ungerr.NotFoundError(fmt.Sprintf("expense bill with ID %s is not found", spec.Model.ID))
		}

		if bill.Status == expenses.FailedExtracting {
			bill.Status = expenses.PendingBill
			if _, err := ebs.billRepo.Update(ctx, bill); err != nil {
				return err
			}
			return ebs.taskQueue.Enqueue(ctx, message.ExpenseBillUploaded{ID: billID})
		}

		if bill.ExtractedText == "" {
			return ungerr.UnprocessableEntityError(fmt.Sprintf("bill %s has no extracted text to be parsed", billID))
		}

		switch bill.Status {
		case expenses.PendingBill:
			return ungerr.UnprocessableEntityError(fmt.Sprintf("bill %s is still pending to be extracted", billID))
		case expenses.FailedExtracting:
			return ungerr.UnprocessableEntityError(fmt.Sprintf("bill %s is failed while extracting text", billID))
		case expenses.NotDetectedBill:
			return ungerr.UnprocessableEntityError(fmt.Sprintf("bill %s already parsed with no detected data", billID))
		}

		bill.Status = expenses.ExtractedBill
		if _, err := ebs.billRepo.Update(ctx, bill); err != nil {
			return err
		}

		return ebs.taskQueue.Enqueue(ctx, message.ExpenseBillTextExtracted{ID: billID})
	})
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

	return ebs.imageSvc.DeleteAllInvalid(ctx, config.Global.BucketNameExpenseBill, validObjectKeys)
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

func ObjectKeyToFileID(objectKey string) storage.FileIdentifier {
	return storage.FileIdentifier{
		BucketName: config.Global.BucketNameExpenseBill,
		ObjectKey:  objectKey,
	}
}
