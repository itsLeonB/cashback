package storage

import (
	"context"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/ungerr"
)

type ImageService interface {
	Upload(ctx context.Context, req *ImageUploadRequest) (string, error)
	GetURL(ctx context.Context, fileID FileIdentifier) (string, error)
	GetURI(fileID FileIdentifier) string
	Delete(ctx context.Context, fileID FileIdentifier) error
	DeleteAllInvalid(ctx context.Context, bucketName string, validObjectKeys []string) error
}

type imageServiceImpl struct {
	validate    *validator.Validate
	storageRepo StorageRepository
}

func NewImageService(
	validate *validator.Validate,
	storageRepo StorageRepository,
) ImageService {
	return &imageServiceImpl{
		validate,
		storageRepo,
	}
}

func (ubs *imageServiceImpl) Upload(ctx context.Context, req *ImageUploadRequest) (string, error) {
	if err := ubs.validateUploadRequest(req); err != nil {
		return "", err
	}

	storageReq := StorageUploadRequest{
		Data:           req.ImageData,
		ContentType:    req.ContentType,
		FileIdentifier: req.FileIdentifier,
	}

	if err := ubs.storageRepo.Upload(ctx, &storageReq); err != nil {
		return "", err
	}

	return ubs.storageRepo.ToURI(storageReq.FileIdentifier), nil
}

func (ubs *imageServiceImpl) GetURL(ctx context.Context, fileID FileIdentifier) (string, error) {
	return ubs.storageRepo.GetSignedURL(ctx, fileID, SignedURLDuration)
}

func (ubs *imageServiceImpl) GetURI(fileID FileIdentifier) string {
	return ubs.storageRepo.ToURI(fileID)
}

func (ubs *imageServiceImpl) Delete(ctx context.Context, fileID FileIdentifier) error {
	return ubs.storageRepo.Delete(ctx, fileID)
}

func (ubs *imageServiceImpl) validateUploadRequest(req *ImageUploadRequest) error {
	if req == nil {
		return ungerr.BadRequestError("request is nil")
	}
	if len(req.ImageData) == 0 {
		return ungerr.BadRequestError("image data is required")
	}
	if len(req.ImageData) > MaxFileSize {
		return ungerr.BadRequestError("file is too large")
	}
	if err := ubs.validate.Struct(req); err != nil {
		return ungerr.Wrap(err, "struct validation failed")
	}
	return nil
}

func (ubs *imageServiceImpl) DeleteAllInvalid(ctx context.Context, bucketName string, validObjectKeys []string) error {
	validKeys := make(map[string]struct{}, len(validObjectKeys))
	for _, key := range validObjectKeys {
		validKeys[key] = struct{}{}
	}

	allObjectKeys, err := ubs.storageRepo.GetAllObjectKeys(ctx, bucketName)
	if err != nil {
		return err
	}

	logger.Infof("obtained object keys from bucket: %s,\n%s", bucketName, strings.Join(allObjectKeys, "\n"))

	hasDeleted := false
	for _, key := range allObjectKeys {
		if _, exists := validKeys[key]; !exists {
			hasDeleted = true
			logger.Infof("deleting image: %s", key)

			if e := ubs.storageRepo.Delete(ctx, FileIdentifier{
				BucketName: bucketName,
				ObjectKey:  key,
			}); e != nil {
				logger.Errorf("error deleting image: %s from bucket: %s: %v", key, bucketName, e)
			}
		}
	}

	if !hasDeleted {
		logger.Info("no images to delete")
	}

	return nil
}
