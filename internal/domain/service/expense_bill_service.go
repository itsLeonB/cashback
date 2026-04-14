package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/otel"
	"github.com/itsLeonB/cashback/internal/core/service/ocr"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type expenseBillServiceImpl struct {
	taskQueue            queue.TaskQueue
	billRepo             repository.ExpenseBillRepository
	transactor           crud.Transactor
	imageSvc             storage.ImageService
	ocrSvc               ocr.OCRService
	expenseSvc           GroupExpenseService
	subscriptionLimitSvc SubscriptionLimitService
}

func NewExpenseBillService(
	taskQueue queue.TaskQueue,
	billRepo repository.ExpenseBillRepository,
	transactor crud.Transactor,
	imageSvc storage.ImageService,
	ocrSvc ocr.OCRService,
	expenseSvc GroupExpenseService,
	subscriptionLimitSvc SubscriptionLimitService,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		taskQueue,
		billRepo,
		transactor,
		imageSvc,
		ocrSvc,
		expenseSvc,
		subscriptionLimitSvc,
	}
}

func (ebs *expenseBillServiceImpl) SavePresigned(ctx context.Context, req dto.PresignedExpenseBillRequest) (dto.PresignedExpenseBillResponse, error) {
	ctx, span := otel.Tracer.Start(ctx, "ExpenseBillService.SavePresigned")
	defer span.End()

	var resp dto.PresignedExpenseBillResponse
	err := ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		bill, err := ebs.getBillForUpload(ctx, req.ProfileID, req.GroupExpenseID)
		if err != nil {
			return err
		}

		if bill.ID == uuid.Nil {
			bill.ImageName = ObjectKeyToFileID(util.GenerateObjectKey(req.Filename)).ObjectKey
		}

		bill.Status = expenses.NotUploadedBill
		var savedBill expenses.ExpenseBill
		if bill.ID == uuid.Nil {
			savedBill, err = ebs.billRepo.Insert(ctx, bill)
		} else {
			savedBill, err = ebs.billRepo.Update(ctx, bill)
		}
		if err != nil {
			return err
		}

		fileID := ObjectKeyToFileID(savedBill.ImageName)
		uploadURL, err := ebs.imageSvc.GetUploadURL(fileID)
		if err != nil {
			return err
		}

		resp.BillID = savedBill.ID
		resp.UploadURL = uploadURL

		return nil
	})
	return resp, err
}

func (ebs *expenseBillServiceImpl) getBillForUpload(ctx context.Context, profileID, expenseID uuid.UUID) (expenses.ExpenseBill, error) {
	if err := ebs.subscriptionLimitSvc.CheckUploadLimit(ctx, profileID); err != nil {
		return expenses.ExpenseBill{}, err
	}

	expense, err := ebs.expenseSvc.GetUnconfirmedForUpdate(ctx, profileID, expenseID)
	if err != nil {
		return expenses.ExpenseBill{}, err
	}

	// No existing bill - nothing to check
	if expense.Bill.IsZero() {
		return expenses.ExpenseBill{GroupExpenseID: expenseID}, nil
	}

	switch expense.Bill.Status {
	case expenses.NotUploadedBill, expenses.FailedExtracting, expenses.FailedParsingBill, expenses.NotDetectedBill:
		return expense.Bill, nil
	case expenses.PendingBill, expenses.ExtractedBill, expenses.ParsedBill:
		return expenses.ExpenseBill{}, ungerr.BadRequestError("Not allowed to reupload")
	default:
		return expenses.ExpenseBill{}, ungerr.UnprocessableEntityError("unknown bill status")
	}
}

func (ebs *expenseBillServiceImpl) ExtractBillText(ctx context.Context, msg message.ExpenseBillUploaded) error {
	ctx, span := otel.Tracer.Start(ctx, "ExpenseBillService.ExtractBillText")
	defer span.End()

	err := ebs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[expenses.ExpenseBill]{}
		spec.Model.ID = msg.ID
		spec.ForUpdate = true
		bill, err := ebs.billRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if bill.IsZero() {
			logger.Errorf("expense bill with ID %s is not found", msg.ID)
			return nil
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
	ctx, span := otel.Tracer.Start(ctx, "ExpenseBillService.TriggerParsing")
	defer span.End()

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
			return ungerr.NotFoundError(fmt.Sprintf("expense bill with ID %s is not found", billID))
		}

		if bill.Status == expenses.FailedExtracting || bill.Status == expenses.NotUploadedBill {
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
	ctx, span := otel.Tracer.Start(ctx, "ExpenseBillService.Cleanup")
	defer span.End()

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

func ObjectKeyToFileID(objectKey string) storage.FileIdentifier {
	return storage.FileIdentifier{
		BucketName: config.Global.BucketNameExpenseBill,
		ObjectKey:  objectKey,
	}
}
