package message

import "github.com/google/uuid"

type ExpenseConfirmed struct {
	ID uuid.UUID `json:"id"`
}

func (ExpenseConfirmed) Type() string {
	return "expense-confirmed"
}
