package dto

import (
	"github.com/google/uuid"
)

type NewExpenseBillRequest struct {
	ImageData      []byte
	ProfileID      uuid.UUID
	GroupExpenseID uuid.UUID
	ContentType    string
	Filename       string
	FileSize       int64
}

type ExpenseBillResponse struct {
	BaseDTO
	ImageURL string `json:"imageUrl"`
	Status   string `json:"status"`
}

type PresignedExpenseBillRequest struct {
	ProfileID      uuid.UUID `json:"-"`
	GroupExpenseID uuid.UUID `json:"-"`
	Filename       string    `json:"fileName" binding:"required,min=3"`
}

type PresignedExpenseBillResponse struct {
	BillID    uuid.UUID `json:"billId"`
	UploadURL string    `json:"uploadUrl"`
}
