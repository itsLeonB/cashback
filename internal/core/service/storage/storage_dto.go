package storage

import "io"

type StorageUploadRequest struct {
	Data         []byte
	Reader       io.Reader
	ContentType  string
	CacheControl string
	FileIdentifier
}
