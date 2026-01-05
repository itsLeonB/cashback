package dto

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
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
	BaseDTO
	TotalAmount      decimal.Decimal        `json:"totalAmount"`
	ItemsTotalAmount decimal.Decimal        `json:"itemsTotalAmount"`
	FeesTotalAmount  decimal.Decimal        `json:"feesTotalAmount"`
	Description      string                 `json:"description"`
	Status           expenses.ExpenseStatus `json:"status"`

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

type NewDraftRequest struct {
	Description string `json:"description"`
}

type ExpenseParticipantsRequest struct {
	ParticipantProfileIDs []uuid.UUID `json:"participantProfileIds" binding:"required,min=1"`
	PayerProfileID        uuid.UUID   `json:"payerProfileId" binding:"required"`
	UserProfileID         uuid.UUID   `json:"-"`
	GroupExpenseID        uuid.UUID   `json:"-"`
}
