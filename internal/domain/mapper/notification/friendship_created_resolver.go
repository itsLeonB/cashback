package notification

import (
	"fmt"

	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/message"
)

type friendshipCreatedResolver struct{}

func (friendshipCreatedResolver) Type() string {
	return "friendship-created"
}

func (friendshipCreatedResolver) ResolveTitle(n entity.Notification) (string, error) {
	metadata, err := unmarshal[message.FriendRequestAcceptedMetadata](n.Metadata)
	if err != nil {
		return "", err
	}

	if metadata.FriendName == "" {
		return "You have a new friend", nil
	}

	return fmt.Sprintf("You are now friends with %s", metadata.FriendName), nil
}
