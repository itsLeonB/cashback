package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/shopspring/decimal"
)

type NewGroupExpenseRequest struct {
	CreatorProfileID uuid.UUID               `json:"-"`
	PayerProfileID   uuid.UUID               `json:"payerProfileId"`
	TotalAmount      decimal.Decimal         `json:"totalAmount" binding:"required"`
	Subtotal         decimal.Decimal         `json:"subtotal" binding:"required"`
	Description      string                  `json:"description"`
	Items            []NewExpenseItemRequest `json:"items" binding:"required,min=1,dive"`
	OtherFees        []NewOtherFeeRequest    `json:"otherFees" binding:"dive"`
}

type GroupExpenseResponse struct {
	ID               uuid.UUID       `json:"id"`
	TotalAmount      decimal.Decimal `json:"totalAmount"`
	ItemsTotalAmount decimal.Decimal `json:"itemsTotalAmount"`
	FeesTotalAmount  decimal.Decimal `json:"feesTotalAmount"`
	Description      string          `json:"description"`
	// Deprecated: refer to Status instead
	Confirmed bool `json:"confirmed"`
	// Deprecated: refer to Status instead
	ParticipantsConfirmed bool                      `json:"participantsConfirmed"`
	Status                appconstant.ExpenseStatus `json:"status"`
	CreatedAt             time.Time                 `json:"createdAt"`
	UpdatedAt             time.Time                 `json:"updatedAt"`
	DeletedAt             time.Time                 `json:"deletedAt,omitzero"`

	// Relationships
	Payer        SimpleProfile                `json:"payer"`
	Creator      SimpleProfile                `json:"creator"`
	Items        []ExpenseItemResponse        `json:"items"`
	OtherFees    []OtherFeeResponse           `json:"otherFees"`
	Participants []ExpenseParticipantResponse `json:"participants"`
	Bill         ExpenseBillResponse          `json:"bill"`
	BillExists   bool                         `json:"billExists"`
}

type ExpenseParticipantResponse struct {
	Profile     SimpleProfile   `json:"profile"`
	ShareAmount decimal.Decimal `json:"shareAmount"`
}

type ExpenseResponseV2 struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt,omitzero"`

	Payer            Participant               `json:"payer"`
	Creator          Participant               `json:"creator"`
	TotalAmount      decimal.Decimal           `json:"totalAmount"`
	ItemsTotalAmount decimal.Decimal           `json:"itemsTotalAmount"`
	FeesTotalAmount  decimal.Decimal           `json:"feesTotalAmount"`
	Description      string                    `json:"description"`
	Status           appconstant.ExpenseStatus `json:"status"`

	// Relationships
	Items        []ExpenseItemResponse        `json:"items"`
	OtherFees    []OtherFeeResponse           `json:"otherFees"`
	Participants []ExpenseParticipantResponse `json:"participants"`
}

type Participant struct {
	ProfileID uuid.UUID `json:"profileId"`
	Name      string    `json:"name"`
	Avatar    string    `json:"avatar"`
	IsUser    bool      `json:"isUser"`
}

type NewDraftRequest struct {
	Description string `json:"description"`
}

type ExpenseParticipantsRequest struct {
	ParticipantProfileIDs []uuid.UUID `json:"participantProfileIds" binding:"required,min=1"`
	PayerProfileID        uuid.UUID   `json:"payerProfileId" binding:"required"`
	UserProfileID         uuid.UUID   `json:"-"`
	GroupExpenseID        uuid.UUID   `json:"-"`
}
