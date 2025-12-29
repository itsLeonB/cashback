package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ItemParticipantResponse struct {
	Profile    SimpleProfile   `json:"profile"`
	ShareRatio decimal.Decimal `json:"shareRatio"`
}

type ExpenseItemResponse struct {
	ID             uuid.UUID                 `json:"id"`
	GroupExpenseID uuid.UUID                 `json:"groupExpenseId"`
	Name           string                    `json:"name"`
	Amount         decimal.Decimal           `json:"amount"`
	Quantity       int                       `json:"quantity"`
	CreatedAt      time.Time                 `json:"createdAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
	DeletedAt      time.Time                 `json:"deletedAt,omitzero"`
	Participants   []ItemParticipantResponse `json:"participants,omitempty"`
}

type UpdateExpenseItemRequest struct {
	UserProfileID  uuid.UUID                `json:"-"`
	ID             uuid.UUID                `json:"-"`
	GroupExpenseID uuid.UUID                `json:"-"`
	Name           string                   `json:"name" binding:"required,min=3"`
	Amount         decimal.Decimal          `json:"amount" binding:"required"`
	Quantity       int                      `json:"quantity" binding:"required,min=1"`
	Participants   []ItemParticipantRequest `json:"participants" binding:"dive"`
}

type ItemParticipantRequest struct {
	ProfileID uuid.UUID       `json:"profileId" binding:"required"`
	Share     decimal.Decimal `json:"share" binding:"required"`
}

type NewExpenseItemRequest struct {
	UserProfileID  uuid.UUID       `json:"-"`
	GroupExpenseID uuid.UUID       `json:"-"`
	Name           string          `json:"name" binding:"required,min=3"`
	Amount         decimal.Decimal `json:"amount" binding:"required"`
	Quantity       int             `json:"quantity" binding:"required,min=1"`
}
