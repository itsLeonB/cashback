package message

import "github.com/google/uuid"

type ExpenseBillUploaded struct {
	ID  uuid.UUID `json:"id"`
	URI string    `json:"uri"`
}

func (ExpenseBillUploaded) Type() string {
	return "expense-bill-uploaded"
}
