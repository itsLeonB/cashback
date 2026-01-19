package dto

import (
	"io"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/domain/entity/expenses"
)

type NewExpenseBillRequest struct {
	ImageReader    io.ReadCloser
	ProfileID      uuid.UUID
	GroupExpenseID uuid.UUID
	ContentType    string
	Filename       string
	FileSize       int64
}

type ExpenseBillResponse struct {
	BaseDTO
	ImageURL string              `json:"imageUrl"`
	Status   expenses.BillStatus `json:"status"`
}
