package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/orcashtrator/internal/domain/friendship"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/mapper"
)

type friendshipRequestServiceImpl struct {
	requestClient friendship.RequestClient
}

func NewFriendshipRequestService(requestClient friendship.RequestClient) FriendshipRequestService {
	return &friendshipRequestServiceImpl{requestClient}
}

func (frs *friendshipRequestServiceImpl) Send(ctx context.Context, userProfileID, friendProfileID uuid.UUID) error {
	return frs.requestClient.Send(ctx, userProfileID, friendProfileID)
}

func (frs *friendshipRequestServiceImpl) GetAllSent(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error) {
	requests, err := frs.requestClient.GetAllSent(ctx, userProfileID)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSlice(requests, mapper.GetFriendshipRequestSimpleMapper(userProfileID)), nil
}

func (frs *friendshipRequestServiceImpl) Cancel(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.requestClient.Cancel(ctx, userProfileID, reqID)
}

func (frs *friendshipRequestServiceImpl) GetAllReceived(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error) {
	requests, err := frs.requestClient.GetAllReceived(ctx, userProfileID)
	if err != nil {
		return nil, err
	}
	return ezutil.MapSlice(requests, mapper.GetFriendshipRequestSimpleMapper(userProfileID)), nil
}

func (frs *friendshipRequestServiceImpl) Ignore(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.requestClient.Ignore(ctx, userProfileID, reqID)
}

func (frs *friendshipRequestServiceImpl) Block(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.requestClient.Block(ctx, userProfileID, reqID)
}

func (frs *friendshipRequestServiceImpl) Unblock(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.requestClient.Unblock(ctx, userProfileID, reqID)
}

func (frs *friendshipRequestServiceImpl) Accept(ctx context.Context, userProfileID, reqID uuid.UUID) (dto.FriendshipResponse, error) {
	friendship, err := frs.requestClient.Accept(ctx, userProfileID, reqID)
	if err != nil {
		return dto.FriendshipResponse{}, err
	}
	return mapper.FriendshipToResponse(friendship), nil
}
