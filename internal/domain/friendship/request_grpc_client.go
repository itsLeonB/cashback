package friendship

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/friendship/v1"
	"github.com/itsLeonB/ezutil/v2"
	"google.golang.org/grpc"
)

type RequestClient interface {
	Send(ctx context.Context, userProfileID, friendProfileID uuid.UUID, message string) error
	GetAllSent(ctx context.Context, userProfileID uuid.UUID) ([]Request, error)
	Cancel(ctx context.Context, userProfileID, reqID uuid.UUID) error
	GetAllReceived(ctx context.Context, userProfileID uuid.UUID) ([]Request, error)
	Ignore(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Block(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Unblock(ctx context.Context, userProfileID, reqID uuid.UUID) error
	Accept(ctx context.Context, userProfileID, reqID uuid.UUID) (Friendship, error)
}

type requestClient struct {
	client friendship.RequestServiceClient
}

func NewRequestClient(conn *grpc.ClientConn) RequestClient {
	return &requestClient{friendship.NewRequestServiceClient(conn)}
}

func (rc *requestClient) Send(ctx context.Context, userProfileID, friendProfileID uuid.UUID, message string) error {
	request := friendship.SendRequest{
		UserProfileId:   userProfileID.String(),
		FriendProfileId: friendProfileID.String(),
		Message:         message,
	}
	_, err := rc.client.Send(ctx, &request)
	return err
}

func (rc *requestClient) GetAllSent(ctx context.Context, userProfileID uuid.UUID) ([]Request, error) {
	request := friendship.GetAllSentRequest{
		UserProfileId: userProfileID.String(),
	}
	response, err := rc.client.GetAllSent(ctx, &request)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSliceWithError(response.GetRequests(), fromRequestProto)
}

func (rc *requestClient) Cancel(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	request := friendship.CancelRequest{
		UserProfileId: userProfileID.String(),
		RequestId:     reqID.String(),
	}
	_, err := rc.client.Cancel(ctx, &request)
	return err
}

func (rc *requestClient) GetAllReceived(ctx context.Context, userProfileID uuid.UUID) ([]Request, error) {
	request := friendship.GetAllReceivedRequest{
		UserProfileId: userProfileID.String(),
	}
	response, err := rc.client.GetAllReceived(ctx, &request)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSliceWithError(response.GetRequests(), fromRequestProto)
}

func (rc *requestClient) Ignore(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	request := friendship.IgnoreRequest{
		UserProfileId: userProfileID.String(),
		RequestId:     reqID.String(),
	}
	_, err := rc.client.Ignore(ctx, &request)
	return err
}

func (rc *requestClient) Block(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	request := friendship.BlockRequest{
		UserProfileId: userProfileID.String(),
		RequestId:     reqID.String(),
	}
	_, err := rc.client.Block(ctx, &request)
	return err
}

func (rc *requestClient) Unblock(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	request := friendship.UnblockRequest{
		UserProfileId: userProfileID.String(),
		RequestId:     reqID.String(),
	}
	_, err := rc.client.Unblock(ctx, &request)
	return err
}

func (rc *requestClient) Accept(ctx context.Context, userProfileID, reqID uuid.UUID) (Friendship, error) {
	request := friendship.AcceptRequest{
		UserProfileId: userProfileID.String(),
		RequestId:     reqID.String(),
	}
	response, err := rc.client.Accept(ctx, &request)
	if err != nil {
		return Friendship{}, err
	}
	return fromFriendshipProto(response.GetFriendship())
}
