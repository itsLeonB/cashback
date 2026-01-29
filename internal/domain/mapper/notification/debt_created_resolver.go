package notification

import (
	"fmt"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/message"
)

type debtCreatedResolver struct{}

func (debtCreatedResolver) Type() string {
	return message.DebtCreated{}.Type()
}

func (debtCreatedResolver) ResolveTitle(n entity.Notification) (string, error) {
	metadata, err := unmarshal[message.DebtCreatedMetadata](n.Metadata)
	if err != nil {
		return "", err
	}

	if metadata.FriendName == "" {
		return "New Transaction", nil
	}

	return fmt.Sprintf("New Transaction with %s", metadata.FriendName), nil
}
