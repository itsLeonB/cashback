package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/itsLeonB/ungerr"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
)

type StorageRepository interface {
	Upload(ctx context.Context, req *StorageUploadRequest) error
	Delete(ctx context.Context, fileID FileIdentifier) error
	GetSignedURL(ctx context.Context, fileID FileIdentifier, expiration time.Duration) (string, error)
	GetAllObjectKeys(ctx context.Context, bucketName string) ([]string, error)
	Exists(ctx context.Context, fileID FileIdentifier) (bool, error)
	ToURI(fi FileIdentifier) string
	Close() error
}

type gcsStorageRepository struct {
	client *storage.Client
}

func NewGCSStorageRepository() (StorageRepository, error) {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, ungerr.Unknownf("failed to create GCS client: %v", err)
	}

	return &gcsStorageRepository{client}, nil
}

func (r *gcsStorageRepository) Upload(ctx context.Context, req *StorageUploadRequest) error {
	bucket := r.client.Bucket(req.BucketName)
	obj := bucket.Object(req.ObjectKey)

	// Create a writer to upload the file
	writer := obj.NewWriter(ctx)
	writer.ContentType = req.ContentType
	writer.Metadata = map[string]string{
		"uploaded_at": time.Now().Format(time.RFC3339),
	}

	// Set cache control for images
	writer.CacheControl = "public, max-age=3600" // 1 hour cache

	// Write the file data
	if _, err := io.Copy(writer, bytes.NewReader(req.Data)); err != nil {
		_ = writer.Close() // best-effort close on copy failure
		return ungerr.Wrap(err, "failed to upload file to GCS")
	}
	if err := writer.Close(); err != nil {
		return ungerr.Wrap(err, "failed to finalize upload to GCS")
	}

	// // Make the object publicly readable (optional)
	// if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
	// 	// Log warning but don't fail the operation
	// 	// In production, you might want to handle this differently based on your security requirements
	// 	fmt.Printf("Warning: failed to make object public: %v\n", err)
	// }

	// // Generate public URL
	// publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", req.BucketName, req.ObjectKey)

	return nil
}

func (r *gcsStorageRepository) Delete(ctx context.Context, fileID FileIdentifier) error {
	if err := r.toObject(fileID).Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			// Object doesn't exist, consider it already deleted
			return nil
		}
		return ungerr.Wrap(err, "failed to delete file from GCS")
	}

	return nil
}

func (r *gcsStorageRepository) GetSignedURL(ctx context.Context, fileID FileIdentifier, expiration time.Duration) (string, error) {
	url, err := r.toBucket(fileID).SignedURL(fileID.ObjectKey, &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  http.MethodGet,
		Expires: time.Now().Add(expiration),
	})
	if err != nil {
		return "", ungerr.Wrap(err, "failed to generate signed URL")
	}

	return url, nil
}

func (r *gcsStorageRepository) GetAllObjectKeys(ctx context.Context, bucketName string) ([]string, error) {
	bucket := r.client.Bucket(bucketName)
	it := bucket.Objects(ctx, nil)
	objectKeys := make([]string, 0)

	for {
		attr, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, ungerr.Wrap(err, "error listing objects in bucket")
		}
		objectKeys = append(objectKeys, attr.Name)
	}

	return objectKeys, nil
}

func (r *gcsStorageRepository) Exists(ctx context.Context, fileID FileIdentifier) (bool, error) {
	_, err := r.toObject(fileID).Attrs(ctx)
	if err == nil {
		return true, nil
	}

	// Case 1: canonical GCS error
	if errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	}

	// Case 2: wrapped googleapi error (most common)
	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		if gErr.Code == http.StatusNotFound {
			return false, nil
		}
	}

	return false, err
}

func (r *gcsStorageRepository) ToURI(fi FileIdentifier) string {
	return fmt.Sprintf("gs://%s/%s", fi.BucketName, fi.ObjectKey)
}

func (r *gcsStorageRepository) Close() error {
	return r.client.Close()
}

func (r *gcsStorageRepository) toObject(fi FileIdentifier) *storage.ObjectHandle {
	return r.client.Bucket(fi.BucketName).Object(fi.ObjectKey)
}

func (r *gcsStorageRepository) toBucket(fi FileIdentifier) *storage.BucketHandle {
	return r.client.Bucket(fi.BucketName)
}
