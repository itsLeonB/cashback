package dto

import (
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type DebtTransactionDirection string

const (
	IncomingDebt DebtTransactionDirection = "INCOMING"
	OutgoingDebt DebtTransactionDirection = "OUTGOING"
)

type NewDebtTransactionRequest struct {
	UserProfileID    uuid.UUID                `json:"-"`
	FriendProfileID  uuid.UUID                `json:"friendProfileId" binding:"required"`
	Direction        DebtTransactionDirection `json:"direction" binding:"oneof=INCOMING OUTGOING"`
	Currency         string                   `json:"currency" binding:"len=3"`
	Amount           decimal.Decimal          `json:"amount" binding:"required"`
	TransferMethodID uuid.UUID                `json:"transferMethodId" binding:"required"`
	Description      string                   `json:"description"`
}

type DebtTransactionResponse struct {
	BaseDTO
	Profile        SimpleProfile   `json:"profile"`
	Type           string          `json:"type"` // "LENT" or "BORROWED"
	Currency       string          `json:"currency"`
	Amount         decimal.Decimal `json:"amount"`
	TransferMethod string          `json:"transferMethod"`
	Description    string          `json:"description"`
	GroupExpenseID uuid.UUID       `json:"groupExpenseId"`
	IsFromExpense  bool            `json:"isFromExpense"`
}
