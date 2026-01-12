package service

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/core/logger"
	"github.com/itsLeonB/cashback/internal/core/service/cache"
	"github.com/itsLeonB/cashback/internal/core/service/storage"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/repository"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type transferMethodServiceImpl struct {
	transferMethodRepo repository.TransferMethodRepository
	storageRepo        storage.StorageRepository
	bucketName         string
	fs                 embed.FS
	urlCache           cache.Cache[string]
}

func NewTransferMethodService(
	transferMethodRepo repository.TransferMethodRepository,
	storageRepo storage.StorageRepository,
	bucketName string,
	fs embed.FS,
) TransferMethodService {
	return &transferMethodServiceImpl{
		transferMethodRepo,
		storageRepo,
		bucketName,
		fs,
		cache.NewInMemoryCache[string](iconURLExpiry),
	}
}

var spaceRegex = regexp.MustCompile(`\s+`)
var iconURLExpiry = 7 * 24 * time.Hour // 7 days

func (tms *transferMethodServiceImpl) GetAll(ctx context.Context, filter debts.ParentFilter, profileID uuid.UUID) ([]dto.TransferMethodResponse, error) {
	methods, err := tms.transferMethodRepo.GetAllByParentFilter(ctx, filter, profileID)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(methods, tms.SignedURLPopulator(ctx)), nil
}

func (tms *transferMethodServiceImpl) SignedURLPopulator(ctx context.Context) func(debts.TransferMethod) dto.TransferMethodResponse {
	return func(tm debts.TransferMethod) dto.TransferMethodResponse {
		if !tm.IconURL.Valid {
			return mapper.TransferMethodToResponse(tm, "")
		}

		url, ok := tms.urlCache.Get(ctx, tm.IconURL.String, tms.getIconURL)
		if !ok {
			return mapper.TransferMethodToResponse(tm, "")
		}

		return mapper.TransferMethodToResponse(tm, url)
	}
}

func (tms *transferMethodServiceImpl) getIconURL(ctx context.Context, objectKey string) (string, bool) {
	url, err := tms.storageRepo.GetSignedURL(
		ctx,
		storage.FileIdentifier{
			BucketName: tms.bucketName,
			ObjectKey:  objectKey,
		},
		iconURLExpiry,
	)
	if err != nil {
		logger.Error(err)
		return "", false
	}
	return url, true
}

func (tms *transferMethodServiceImpl) GetByID(ctx context.Context, id uuid.UUID) (debts.TransferMethod, error) {
	spec := crud.Specification[debts.TransferMethod]{}
	spec.Model.ID = id

	transferMethod, err := tms.transferMethodRepo.FindFirst(ctx, spec)
	if err != nil {
		return debts.TransferMethod{}, err
	}
	if transferMethod.IsZero() {
		return debts.TransferMethod{}, ungerr.NotFoundError(fmt.Sprintf(appconstant.ErrTransferMethodNotFound, id))
	}

	return transferMethod, nil
}

func (tms *transferMethodServiceImpl) GetByName(ctx context.Context, name string) (debts.TransferMethod, error) {
	spec := crud.Specification[debts.TransferMethod]{}
	spec.Model.Name = name

	transferMethod, err := tms.transferMethodRepo.FindFirst(ctx, spec)
	if err != nil {
		return debts.TransferMethod{}, err
	}
	if transferMethod.IsZero() {
		return debts.TransferMethod{}, ungerr.Unknownf("%s transfer method not found", name)
	}

	return transferMethod, nil
}

func (tms *transferMethodServiceImpl) SyncMethods(ctx context.Context) error {
	newMethods := []debts.TransferMethod{}
	parents, existingChildren, err := tms.prepareForMethodSync(ctx)
	if err != nil {
		return err
	}

	for _, parent := range parents {
		logger.Infof("reading for parent transfer method: %s", parent.Name)
		entries, err := tms.fs.ReadDir(path.Join("assets/transfer-methods", parent.Name))
		if err != nil {
			// Skip if directory doesn't exist in embedded filesystem
			logger.Warnf("directory not found in embedded filesystem: %s, skipping...", parent.Name)
			continue
		}

		for _, e := range entries {
			name, display, skip := returnNameDisplayOrSkip(e)
			if skip {
				continue
			}

			logger.Infof("found filename %s", e.Name())
			if _, exists := existingChildren[name]; exists {
				logger.Infof("transfer method %s already exists in database, skipping...", name)
				continue
			}

			fileID := storage.FileIdentifier{
				BucketName: tms.bucketName,
				ObjectKey:  buildIconKey(parent.Name, name),
			}

			if err = tms.uploadIcon(ctx, parent.Name, e.Name(), fileID); err != nil {
				return err
			}

			newMethods = append(newMethods, debts.TransferMethod{
				Name:    name,
				Display: display,
				IconURL: sql.NullString{
					String: fileID.ObjectKey,
					Valid:  true,
				},
				ParentID: uuid.NullUUID{
					UUID:  parent.ID,
					Valid: true,
				},
			})
		}
	}

	if len(newMethods) < 1 {
		return nil
	}

	_, err = tms.transferMethodRepo.SaveMany(ctx, newMethods)
	return err
}

func (tms *transferMethodServiceImpl) prepareForMethodSync(ctx context.Context) ([]debts.TransferMethod, map[string]struct{}, error) {
	methods, err := tms.transferMethodRepo.FindAll(ctx, crud.Specification[debts.TransferMethod]{})
	if err != nil {
		return nil, nil, err
	}

	parents := make([]debts.TransferMethod, 0, len(methods))
	existingChildren := make(map[string]struct{}, len(methods))

	for _, method := range methods {
		if method.ParentID.Valid {
			existingChildren[method.Name] = struct{}{}
		} else {
			parents = append(parents, method)
		}
	}

	return parents, existingChildren, nil
}

func (tms *transferMethodServiceImpl) uploadIcon(ctx context.Context, parentName, fileName string, fileID storage.FileIdentifier) error {
	iconExists, err := tms.storageRepo.Exists(ctx, fileID)
	if err != nil {
		return err
	}
	if iconExists {
		logger.Infof("%s already uploaded, skipping upload...", fileName)
		return nil
	}

	f, err := tms.fs.Open(path.Join("assets/transfer-methods", parentName, fileName))
	if err != nil {
		return ungerr.Wrap(err, "error opening embedded filesystem")
	}

	defer func() {
		if cerr := f.Close(); cerr != nil {
			logger.Errorf("error closing file: %v", cerr)
		}
	}()

	if err = tms.storageRepo.Upload(ctx, &storage.StorageUploadRequest{
		FileIdentifier: fileID,
		Reader:         f,
		ContentType:    "image/svg+xml",
		CacheControl:   "public, max-age=31536000, immutable",
	}); err != nil {
		return err
	}

	logger.Infof("success uploading %s", fileName)
	return nil
}

func returnNameDisplayOrSkip(e fs.DirEntry) (string, string, bool) {
	if e.IsDir() {
		return "", "", true
	}
	if !strings.HasSuffix(e.Name(), ".svg") {
		return "", "", true
	}

	display := extractDisplayFromFileName(e.Name())
	name := normalizeName(display)

	return name, display, false
}

func extractDisplayFromFileName(fileName string) string {
	fileName = strings.TrimSuffix(fileName, ".svg")
	fileName = strings.TrimSpace(strings.SplitN(fileName, "(", 2)[0])
	return fileName
}

func normalizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = spaceRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

func buildIconKey(parentName, name string) string {
	return fmt.Sprintf("%s/%s.svg", parentName, name)
}
