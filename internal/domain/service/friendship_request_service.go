package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/core/service/queue"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/entity"
	"github.com/itsLeonB/cashback/internal/domain/entity/users"
	"github.com/itsLeonB/cashback/internal/domain/mapper"
	"github.com/itsLeonB/cashback/internal/domain/message"
	"github.com/itsLeonB/ezutil/v2"
	"github.com/itsLeonB/go-crud"
	"github.com/itsLeonB/ungerr"
)

type friendshipRequestServiceImpl struct {
	transactor     crud.Transactor
	friendshipSvc  FriendshipService
	profileService ProfileService
	requestRepo    crud.Repository[users.FriendshipRequest]
	taskQueue      queue.TaskQueue
}

func NewFriendshipRequestService(
	transactor crud.Transactor,
	friendshipSvc FriendshipService,
	profileService ProfileService,
	requestRepo crud.Repository[users.FriendshipRequest],
	taskQueue queue.TaskQueue,
) FriendshipRequestService {
	return &friendshipRequestServiceImpl{
		transactor,
		friendshipSvc,
		profileService,
		requestRepo,
		taskQueue,
	}
}

func (frs *friendshipRequestServiceImpl) Send(ctx context.Context, userProfileID, friendProfileID uuid.UUID) error {
	return frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.SenderProfileID = userProfileID
		spec.Model.RecipientProfileID = friendProfileID
		existingRequest, err := frs.requestRepo.FindFirst(ctx, spec)
		if err != nil {
			return err
		}
		if !existingRequest.IsZero() {
			if existingRequest.BlockedAt.Valid {
				return ungerr.UnprocessableEntityError("user is blocked by recipient")
			}
			return ungerr.UnprocessableEntityError("user still has existing request")
		}

		if err = frs.validateFriendProfile(ctx, userProfileID, friendProfileID); err != nil {
			return err
		}

		insertedRequest, err := frs.requestRepo.Insert(ctx, users.FriendshipRequest{
			SenderProfileID:    userProfileID,
			RecipientProfileID: friendProfileID,
		})
		if err != nil {
			return err
		}

		go frs.taskQueue.AsyncEnqueue(message.FriendRequestSent{ID: insertedRequest.ID})
		return nil
	})
}

func (frs *friendshipRequestServiceImpl) validateFriendProfile(ctx context.Context, userProfileID, friendProfileID uuid.UUID) error {
	isFriends, _, err := frs.friendshipSvc.IsFriends(ctx, userProfileID, friendProfileID)
	if err != nil {
		return err
	}
	if isFriends {
		return ungerr.UnprocessableEntityError("already friends")
	}
	friendProfile, err := frs.profileService.GetByID(ctx, friendProfileID)
	if err != nil {
		return err
	}
	if friendProfile.UserID == uuid.Nil {
		return ungerr.UnprocessableEntityError("cannot request friendship with anonymous profile")
	}
	return nil
}

func (frs *friendshipRequestServiceImpl) GetAllSent(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error) {
	spec := crud.Specification[users.FriendshipRequest]{}
	spec.Model.SenderProfileID = userProfileID
	spec.PreloadRelations = []string{"SenderProfile", "RecipientProfile"}
	requests, err := frs.requestRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	response := make([]dto.FriendshipRequestResponse, 0, len(requests))
	for _, request := range requests {
		if request.BlockedAt.Valid {
			continue
		}
		response = append(response, mapper.FriendshipRequestToResponse(request, userProfileID))
	}

	return response, nil
}

func (frs *friendshipRequestServiceImpl) Cancel(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = reqID
		spec.Model.SenderProfileID = userProfileID
		spec.ForUpdate = true
		request, err := frs.getPendingRequest(ctx, spec)
		if err != nil {
			return err
		}
		return frs.requestRepo.Delete(ctx, request)
	})
}

func (frs *friendshipRequestServiceImpl) getPendingRequest(ctx context.Context, spec crud.Specification[users.FriendshipRequest]) (users.FriendshipRequest, error) {
	request, err := frs.getRequest(ctx, spec)
	if err != nil {
		return users.FriendshipRequest{}, err
	}
	if request.BlockedAt.Valid {
		return users.FriendshipRequest{}, ungerr.UnprocessableEntityError("sender is blocked")
	}
	return request, nil
}

func (frs *friendshipRequestServiceImpl) getRequest(ctx context.Context, spec crud.Specification[users.FriendshipRequest]) (users.FriendshipRequest, error) {
	request, err := frs.requestRepo.FindFirst(ctx, spec)
	if err != nil {
		return users.FriendshipRequest{}, err
	}
	if request.IsZero() {
		return users.FriendshipRequest{}, ungerr.NotFoundError("request not found")
	}
	return request, nil
}

func (frs *friendshipRequestServiceImpl) GetAllReceived(ctx context.Context, userProfileID uuid.UUID) ([]dto.FriendshipRequestResponse, error) {
	spec := crud.Specification[users.FriendshipRequest]{}
	spec.Model.RecipientProfileID = userProfileID
	spec.PreloadRelations = []string{"SenderProfile", "RecipientProfile"}
	requests, err := frs.requestRepo.FindAll(ctx, spec)
	if err != nil {
		return nil, err
	}

	return ezutil.MapSlice(requests, mapper.GetFriendshipRequestSimpleMapper(userProfileID)), nil
}

func (frs *friendshipRequestServiceImpl) Ignore(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = reqID
		spec.Model.RecipientProfileID = userProfileID
		spec.ForUpdate = true
		request, err := frs.getPendingRequest(ctx, spec)
		if err != nil {
			return err
		}
		return frs.requestRepo.Delete(ctx, request)
	})
}

func (frs *friendshipRequestServiceImpl) Block(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = reqID
		spec.Model.RecipientProfileID = userProfileID
		spec.ForUpdate = true
		request, err := frs.getRequest(ctx, spec)
		if err != nil {
			return err
		}
		if request.BlockedAt.Valid {
			return nil
		}

		request.BlockedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
		_, err = frs.requestRepo.Update(ctx, request)
		return err
	})
}

func (frs *friendshipRequestServiceImpl) Unblock(ctx context.Context, userProfileID, reqID uuid.UUID) error {
	return frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = reqID
		spec.Model.RecipientProfileID = userProfileID
		spec.ForUpdate = true
		request, err := frs.getRequest(ctx, spec)
		if err != nil {
			return err
		}
		if !request.BlockedAt.Valid {
			return nil
		}

		request.BlockedAt = sql.NullTime{}
		_, err = frs.requestRepo.Update(ctx, request)
		return err
	})
}

func (frs *friendshipRequestServiceImpl) Accept(ctx context.Context, userProfileID, reqID uuid.UUID) (dto.FriendshipResponse, error) {
	var response dto.FriendshipResponse
	err := frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = reqID
		spec.Model.RecipientProfileID = userProfileID
		spec.ForUpdate = true
		request, err := frs.getPendingRequest(ctx, spec)
		if err != nil {
			return err
		}

		response, err = frs.friendshipSvc.CreateReal(ctx, userProfileID, request.SenderProfileID)
		if err != nil {
			return err
		}

		return frs.requestRepo.Delete(ctx, request)
	})
	return response, err
}

func (frs *friendshipRequestServiceImpl) ConstructNotification(ctx context.Context, msg message.FriendRequestSent) (entity.Notification, error) {
	var notification entity.Notification
	err := frs.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		spec := crud.Specification[users.FriendshipRequest]{}
		spec.Model.ID = msg.ID
		spec.ForUpdate = true
		req, err := frs.getPendingRequest(ctx, spec)
		if err != nil {
			return err
		}

		notification = entity.Notification{
			ProfileID:  req.RecipientProfileID,
			Type:       "friend-request-received",
			EntityType: "friend-request",
			EntityID:   req.ID,
		}

		return err
	})
	return notification, err
}
