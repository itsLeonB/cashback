package notification

import (
	"fmt"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/ezutil/v2"
)

type debtCreatedResolver struct{}

func (debtCreatedResolver) Type() string {
	return message.DebtCreated{}.Type()
}

func (debtCreatedResolver) ResolveTitle(n entity.Notification) (string, error) {
	metadata, err := ezutil.Unmarshal[message.DebtCreatedMetadata](n.Metadata)
	if err != nil {
		return "", err
	}

	if metadata.FriendName == "" {
		return "New Transaction", nil
	}

	return fmt.Sprintf("New Transaction with %s", metadata.FriendName), nil
}
