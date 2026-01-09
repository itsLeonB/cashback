package message

import "github.com/google/uuid"

type ExpenseBillUploaded struct {
	ID uuid.UUID `json:"id"`
}

func (ExpenseBillUploaded) Type() string {
	return "expense-bill-uploaded"
}
