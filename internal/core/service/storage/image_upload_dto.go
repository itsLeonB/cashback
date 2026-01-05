package storage

type ImageUploadRequest struct {
	ImageData   []byte `validate:"required"`
	ContentType string `validate:"required,oneof=image/jpeg image/png image/jpg image/webp"`
	FileSize    int64  `validate:"required,min=1"`
	FileIdentifier
}

type FileIdentifier struct {
	BucketName string `validate:"required"`
	ObjectKey  string `validate:"required,min=4"`
}
