package notification

import "github.com/itsLeonB/cashback/internal/domain/entity"

type friendRequestReceivedResolver struct{}

func (friendRequestReceivedResolver) Type() string {
	return "friend-request-received"
}

func (friendRequestReceivedResolver) ResolveTitle(n entity.Notification) (string, error) {
	return "New Friend Request", nil
}
