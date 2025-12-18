package service

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/meq"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/config"
	"github.com/itsLeonB/orcashtrator/internal/domain/expensebill"
	"github.com/itsLeonB/orcashtrator/internal/domain/imageupload"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
	"github.com/itsLeonB/orcashtrator/internal/message"
	"github.com/itsLeonB/orcashtrator/internal/util"
	"github.com/itsLeonB/ungerr"
)

type expenseBillServiceImpl struct {
	logger            ezutil.Logger
	friendshipService FriendshipService
	profileService    ProfileService
	expenseBillClient expensebill.ExpenseBillClient
	imageUploadClient imageupload.ImageUploadClient
	bucketName        string
	taskQueue         meq.TaskQueue[message.ExpenseBillUploaded]
}

func NewExpenseBillService(
	logger ezutil.Logger,
	friendshipService FriendshipService,
	profileService ProfileService,
	expenseBillClient expensebill.ExpenseBillClient,
	imageUploadClient imageupload.ImageUploadClient,
	bucketName string,
	taskQueue meq.TaskQueue[message.ExpenseBillUploaded],
) ExpenseBillService {
	return &expenseBillServiceImpl{
		logger,
		friendshipService,
		profileService,
		expenseBillClient,
		imageUploadClient,
		bucketName,
		taskQueue,
	}
}

func (ebs *expenseBillServiceImpl) Save(ctx context.Context, req *dto.NewExpenseBillRequest) (dto.ExpenseBillResponse, error) {
	// Default PayerProfileID to the user's profile ID if not provided
	// This is useful when the user is creating a group expense for themselves.
	if req.PayerProfileID == uuid.Nil {
		req.PayerProfileID = req.CreatorProfileID
	}

	namesByProfileIDs, err := ebs.validateAndGetNames(ctx, req.PayerProfileID, req.CreatorProfileID)
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	savedBill, err := ebs.uploadAndSave(ctx, req)
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	return mapper.ExpenseBillToResponse(savedBill, "", req.CreatorProfileID, namesByProfileIDs), nil
}

func (ebs *expenseBillServiceImpl) SaveV2(ctx context.Context, req *dto.NewExpenseBillRequest) (dto.ExpenseBillResponse, error) {
	savedBill, err := ebs.uploadAndSave(ctx, req)
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	return mapper.ExpenseBillToResponse(savedBill, "", req.CreatorProfileID, make(map[uuid.UUID]string)), nil
}

func (ebs *expenseBillServiceImpl) GetAllCreated(ctx context.Context, creatorProfileID uuid.UUID) ([]dto.ExpenseBillResponse, error) {
	bills, err := ebs.expenseBillClient.GetAllCreated(ctx, creatorProfileID)
	if err != nil {
		return nil, err
	}

	profileIDs := mapper.UniqueBillProfileIDs(bills)
	namesByProfileIDs, err := ebs.profileService.GetNames(ctx, profileIDs)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(bills, mapper.ExpenseBillSimpleMapper("", creatorProfileID, namesByProfileIDs)), nil
}

func (ebs *expenseBillServiceImpl) Get(ctx context.Context, profileID, id uuid.UUID) (dto.ExpenseBillResponse, error) {
	bill, err := ebs.expenseBillClient.Get(ctx, id)
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	imageURL, err := ebs.imageUploadClient.GetURL(ctx, ebs.objectKeyToFileID(bill.ObjectKey))
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	namesByProfileIDs, err := ebs.profileService.GetNames(ctx, []uuid.UUID{bill.CreatorProfileID, bill.PayerProfileID})
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	return mapper.ExpenseBillToResponse(bill, imageURL, profileID, namesByProfileIDs), nil
}

func (ebs *expenseBillServiceImpl) MapToURL(ctx context.Context, bill expensebill.ExpenseBill) (dto.ExpenseBillResponse, error) {
	imageURL, err := ebs.imageUploadClient.GetURL(ctx, ebs.objectKeyToFileID(bill.ObjectKey))
	if err != nil {
		return dto.ExpenseBillResponse{}, err
	}

	return mapper.ExpenseBillToResponse(bill, imageURL, uuid.Nil, make(map[uuid.UUID]string)), nil
}

func (ebs *expenseBillServiceImpl) Delete(ctx context.Context, profileID, id uuid.UUID) error {
	bill, err := ebs.expenseBillClient.Get(ctx, id)
	if err != nil {
		return err
	}

	if err = ebs.expenseBillClient.Delete(ctx, profileID, id); err != nil {
		return err
	}

	if err = ebs.imageUploadClient.Delete(ctx, ebs.objectKeyToFileID(bill.ObjectKey)); err != nil {
		ebs.logger.Errorf("error deleting bill image from storage: %v", err)
	}

	return nil
}

func (ebs *expenseBillServiceImpl) uploadAndSave(ctx context.Context, req *dto.NewExpenseBillRequest) (expensebill.ExpenseBill, error) {
	fileID := ebs.objectKeyToFileID(util.GenerateObjectKey(req.Filename))
	var savedBill expensebill.ExpenseBill
	var wg sync.WaitGroup
	var uri string
	var err error

	wg.Go(func() {
		if billUri, e := ebs.doUpload(ctx, req, fileID); e != nil {
			err = errors.Join(err, e)
		} else {
			uri = billUri
		}
	})

	wg.Go(func() {
		if insertedBill, e := ebs.saveEntry(ctx, req, fileID.ObjectKey); e != nil {
			err = errors.Join(err, e)
		} else {
			savedBill = insertedBill
		}
	})

	wg.Wait()

	if err == nil && uri != "" {
		if e := ebs.taskQueue.Enqueue(ctx, config.AppName, message.ExpenseBillUploaded{
			ID:  savedBill.ID,
			URI: uri,
		}); e != nil {
			err = errors.Join(err, e)
		}
	}

	if err != nil {
		ebs.rollbackUpload(ctx, fileID, req.CreatorProfileID, savedBill.ID)
		return expensebill.ExpenseBill{}, err
	}

	return savedBill, nil
}

func (ebs *expenseBillServiceImpl) rollbackUpload(
	ctx context.Context,
	fileID imageupload.FileIdentifier,
	creatorProfileID, billID uuid.UUID,
) {
	var wg sync.WaitGroup
	wg.Go(func() {
		if err := ebs.imageUploadClient.Delete(ctx, fileID); err != nil {
			ebs.logger.Errorf("error rolling back bill upload: %v", err)
		}
	})
	wg.Go(func() {
		if err := ebs.expenseBillClient.Delete(ctx, creatorProfileID, billID); err != nil {
			ebs.logger.Errorf("error rolling back bill upload: %v", err)
		}
	})
	wg.Wait()
}

func (ebs *expenseBillServiceImpl) doUpload(
	ctx context.Context,
	req *dto.NewExpenseBillRequest,
	fileID imageupload.FileIdentifier,
) (string, error) {
	uploadReq := imageupload.ImageUploadRequest{
		FileStream:     req.ImageReader,
		ContentType:    req.ContentType,
		FileSize:       req.FileSize,
		FileIdentifier: fileID,
	}

	return ebs.imageUploadClient.UploadStream(ctx, &uploadReq)
}

func (ebs *expenseBillServiceImpl) saveEntry(
	ctx context.Context,
	req *dto.NewExpenseBillRequest,
	objectKey string,
) (expensebill.ExpenseBill, error) {
	request := expensebill.ExpenseBill{
		CreatorProfileID: req.CreatorProfileID,
		PayerProfileID:   req.PayerProfileID,
		GroupExpenseID:   req.GroupExpenseID,
		ObjectKey:        objectKey,
	}

	return ebs.expenseBillClient.Save(ctx, request)
}

func (ebs *expenseBillServiceImpl) validateAndGetNames(ctx context.Context, payerProfileID, creatorProfileID uuid.UUID) (map[uuid.UUID]string, error) {
	if payerProfileID == creatorProfileID {
		creatorProfile, err := ebs.profileService.GetByID(ctx, creatorProfileID)
		if err != nil {
			return nil, err
		}

		return map[uuid.UUID]string{creatorProfileID: creatorProfile.Name}, nil
	}

	// Check if the payer is a friend of the user
	if isFriend, _, err := ebs.friendshipService.IsFriends(ctx, creatorProfileID, payerProfileID); err != nil {
		return nil, err
	} else if !isFriend {
		return nil, ungerr.UnprocessableEntityError(appconstant.ErrNotFriends)
	}

	return ebs.profileService.GetNames(ctx, []uuid.UUID{creatorProfileID, payerProfileID})
}

func (ebs *expenseBillServiceImpl) objectKeyToFileID(objectKey string) imageupload.FileIdentifier {
	return imageupload.FileIdentifier{
		BucketName: ebs.bucketName,
		ObjectKey:  objectKey,
	}
}
