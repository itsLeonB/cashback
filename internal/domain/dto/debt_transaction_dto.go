package dto

import (
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/debts"
	"github.com/shopspring/decimal"
)

type NewDebtTransactionRequest struct {
	UserProfileID    uuid.UUID                   `json:"-"`
	FriendProfileID  uuid.UUID                   `json:"friendProfileId" binding:"required"`
	Action           debts.DebtTransactionAction `json:"action" binding:"oneof=LEND BORROW RECEIVE RETURN"`
	Amount           decimal.Decimal             `json:"amount" binding:"required"`
	TransferMethodID uuid.UUID                   `json:"transferMethodId" binding:"required"`
	Description      string                      `json:"description"`
}

type DebtTransactionResponse struct {
	BaseDTO
	ProfileID      uuid.UUID                   `json:"profileId"`
	Type           debts.DebtTransactionType   `json:"type"`
	Action         debts.DebtTransactionAction `json:"action"`
	Amount         decimal.Decimal             `json:"amount"`
	TransferMethod string                      `json:"transferMethod"`
	Description    string                      `json:"description"`
}
