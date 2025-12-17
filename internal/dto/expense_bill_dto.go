package dto

import (
	"io"
	"time"

	"github.com/google/uuid"
)

type NewExpenseBillRequest struct {
	ImageReader      io.ReadCloser
	CreatorProfileID uuid.UUID
	// Deprecated: will be deferred to other API
	PayerProfileID uuid.UUID
	GroupExpenseID uuid.UUID
	ContentType    string
	Filename       string
	FileSize       int64
}

type ExpenseBillResponse struct {
	ID                 uuid.UUID `json:"id"`
	CreatorProfileID   uuid.UUID `json:"creatorProfileId"`
	PayerProfileID     uuid.UUID `json:"payerProfileId"`
	ImageURL           string    `json:"imageUrl"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
	DeletedAt          time.Time `json:"deletedAt,omitzero"`
	IsCreatedByUser    bool      `json:"isCreatedByUser"`
	IsPaidByUser       bool      `json:"isPaidByUser"`
	CreatorProfileName string    `json:"creatorProfileName"`
	PayerProfileName   string    `json:"payerProfileName"`
}
