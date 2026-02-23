package dto

import (
	"time"
)

type SubscriptionResponse struct {
	Plan   string `json:"plan"`
	Limits Limits `json:"limits"`
}

type Limits struct {
	Uploads UploadLimits `json:"uploads"`
}

type UploadLimits struct {
	Daily     UploadLimit `json:"daily"`
	Monthly   UploadLimit `json:"monthly"`
	CanUpload bool        `json:"canUpload"`
}

type UploadLimit struct {
	Used      int       `json:"used"`
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"resetAt"`
	CanUpload bool      `json:"canUpload"`
}
