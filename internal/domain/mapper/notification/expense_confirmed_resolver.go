package notification

import (
	"fmt"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/ezutil/v2"
)

type expenseConfirmedResolver struct{}

func (expenseConfirmedResolver) Type() string {
	return message.ExpenseConfirmed{}.Type()
}

func (expenseConfirmedResolver) ResolveTitle(n entity.Notification) (string, error) {
	metadata, err := ezutil.Unmarshal[message.ExpenseConfirmedMetadata](n.Metadata)
	if err != nil {
		return "", err
	}

	if metadata.CreatorName == "" {
		return "Your friend confirmed an expense with you", nil
	}

	return fmt.Sprintf("%s confirmed an expense with you", metadata.CreatorName), nil
}
