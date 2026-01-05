package storage

type StorageUploadRequest struct {
	Data        []byte
	ContentType string
	FileIdentifier
}
