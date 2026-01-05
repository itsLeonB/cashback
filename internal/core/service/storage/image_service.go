package storage

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/itsLeonB/ungerr"
)

type ImageService interface {
	Upload(ctx context.Context, req *ImageUploadRequest) (string, error)
	GetURL(ctx context.Context, fileID FileIdentifier) (string, error)
	Delete(ctx context.Context, fileID FileIdentifier) error
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
