package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type ItemParticipantResponse struct {
	Profile         SimpleProfile   `json:"profile"`
	Weight          int             `json:"weight"`
	AllocatedAmount decimal.Decimal `json:"allocatedAmount"`
}

type ExpenseItemResponse struct {
	BaseDTO
	GroupExpenseID uuid.UUID                 `json:"groupExpenseId"`
	Name           string                    `json:"name"`
	Amount         decimal.Decimal           `json:"amount"`
	Quantity       int                       `json:"quantity"`
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
	ProfileID uuid.UUID `json:"profileId" binding:"required"`
	Weight    int       `json:"weight"`
}

type NewExpenseItemRequest struct {
	UserProfileID  uuid.UUID       `json:"-"`
	GroupExpenseID uuid.UUID       `json:"-"`
	Name           string          `json:"name" binding:"required,min=3"`
	Amount         decimal.Decimal `json:"amount" binding:"required"`
	Quantity       int             `json:"quantity" binding:"required,min=1"`
}

type SyncItemParticipantsRequest struct {
	ProfileID      uuid.UUID                `json:"-"`
	ID             uuid.UUID                `json:"-"`
	GroupExpenseID uuid.UUID                `json:"-"`
	Participants   []ItemParticipantRequest `json:"participants" binding:"dive"`
}
