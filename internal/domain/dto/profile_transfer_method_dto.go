package dto

import "github.com/google/uuid"

type NewProfileTransferMethodRequest struct {
	ProfileID        uuid.UUID `json:"-"`
	TransferMethodID uuid.UUID `json:"transferMethodId" binding:"required"`
	AccountName      string    `json:"accountName" binding:"required,min=3"`
	AccountNumber    string    `json:"accountNumber" binding:"required,min=3"`
}

type ProfileTransferMethodResponse struct {
	BaseDTO
	Method        TransferMethodResponse `json:"method"`
	AccountName   string                 `json:"accountName"`
	AccountNumber string                 `json:"accountNumber"`
}
