package message

import "github.com/google/uuid"

type ExpenseBillTextExtracted struct {
	ID   uuid.UUID `json:"id"`
	Text string    `json:"text"`
}

func (ExpenseBillTextExtracted) Type() string {
	return "expense-bill-text-extracted"
}
