package service

import (
	"context"
	"io"

	"github.com/itsLeonB/cashback/internal/core/config"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/core/util"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
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
}

func NewExpenseBillService(
	bucketName string,
	taskQueue meq.TaskQueue[message.ExpenseBillUploaded],
	billRepo crud.Repository[expenses.ExpenseBill],
	transactor crud.Transactor,
	imageSvc storage.ImageService,
) ExpenseBillService {
	return &expenseBillServiceImpl{
		bucketName,
		taskQueue,
		billRepo,
		transactor,
		imageSvc,
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
