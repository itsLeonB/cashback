package friendship

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/itsLeonB/cocoon-protos/gen/go/friendship/v1"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/rotisserie/eris"
	"google.golang.org/grpc"
)

type FriendshipClient interface {
	CreateAnonymous(ctx context.Context, req CreateAnonymousRequest) (Friendship, error)
	GetAll(ctx context.Context, profileID uuid.UUID) ([]Friendship, error)
	GetDetails(ctx context.Context, req GetDetailsRequest) (FriendDetails, error)
	IsFriends(ctx context.Context, req IsFriendsRequest) (IsFriendsResponse, error)
}

type friendshipClient struct {
	validate *validator.Validate
	client   friendship.FriendshipServiceClient
}

func NewFriendshipClient(validate *validator.Validate, conn *grpc.ClientConn) FriendshipClient {
	if validate == nil {
		panic("validator is nil")
	}

	return &friendshipClient{
		validate,
		friendship.NewFriendshipServiceClient(conn),
	}
}

func (fc *friendshipClient) CreateAnonymous(ctx context.Context, req CreateAnonymousRequest) (Friendship, error) {
	if err := fc.validate.Struct(req); err != nil {
		return Friendship{}, err
	}

	request := &friendship.CreateAnonymousRequest{
		ProfileId: req.ProfileID.String(),
		Name:      req.Name,
	}

	response, err := fc.client.CreateAnonymous(ctx, request)
	if err != nil {
		return Friendship{}, err
	}

	return fromFriendshipProto(response.GetFriendship())
}

func (fc *friendshipClient) GetAll(ctx context.Context, profileID uuid.UUID) ([]Friendship, error) {
	if profileID == uuid.Nil {
		return nil, eris.New("profileID is nil")
	}

	request := &friendship.GetAllRequest{
		ProfileId: profileID.String(),
	}

	response, err := fc.client.GetAll(ctx, request)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSliceWithError(response.GetFriendships(), fromFriendshipProto)
}

func (fc *friendshipClient) GetDetails(ctx context.Context, req GetDetailsRequest) (FriendDetails, error) {
	if err := fc.validate.Struct(req); err != nil {
		return FriendDetails{}, err
	}

	request := &friendship.GetDetailsRequest{
		ProfileId:    req.ProfileID.String(),
		FriendshipId: req.FriendshipID.String(),
	}

	response, err := fc.client.GetDetails(ctx, request)
	if err != nil {
		return FriendDetails{}, err
	}

	id, err := ezutil.Parse[uuid.UUID](response.GetId())
	if err != nil {
		return FriendDetails{}, err
	}

	profileID, err := ezutil.Parse[uuid.UUID](response.GetProfileId())
	if err != nil {
		return FriendDetails{}, err
	}

	friendshipType, err := fromFriendshipTypeProto(response.GetType())
	if err != nil {
		return FriendDetails{}, err
	}

	profileID1, err := ezutil.Parse[uuid.UUID](response.GetProfileId1())
	if err != nil {
		return FriendDetails{}, err
	}

	profileID2, err := ezutil.Parse[uuid.UUID](response.GetProfileId2())
	if err != nil {
		return FriendDetails{}, err
	}

	return FriendDetails{
		ID:         id,
		ProfileID:  profileID,
		Name:       response.GetName(),
		Type:       friendshipType,
		Email:      response.GetEmail(),
		Phone:      response.GetPhone(),
		Avatar:     response.GetAvatar(),
		CreatedAt:  ezutil.FromProtoTime(response.GetCreatedAt()),
		UpdatedAt:  ezutil.FromProtoTime(response.GetUpdatedAt()),
		DeletedAt:  ezutil.FromProtoTime(response.GetDeletedAt()),
		ProfileID1: profileID1,
		ProfileID2: profileID2,
	}, nil
}

func (fc *friendshipClient) IsFriends(ctx context.Context, req IsFriendsRequest) (IsFriendsResponse, error) {
	if err := fc.validate.Struct(req); err != nil {
		return IsFriendsResponse{}, err
	}

	request := &friendship.IsFriendsRequest{
		ProfileId_1: req.ProfileID1.String(),
		ProfileId_2: req.ProfileID2.String(),
	}

	response, err := fc.client.IsFriends(ctx, request)
	if err != nil {
		return IsFriendsResponse{}, err
	}

	return IsFriendsResponse{
		IsFriends:   response.GetIsFriends(),
		IsAnonymous: response.GetIsAnonymous(),
	}, nil
}
