package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type FriendshipRequestHandler struct {
	svc service.FriendshipRequestService
}

func NewFriendshipRequestHandler(svc service.FriendshipRequestService) *FriendshipRequestHandler {
	return &FriendshipRequestHandler{svc}
}

func (frh *FriendshipRequestHandler) HandleSend() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		friendProfileID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Send(ctx, userProfileID, friendProfileID)
	})
}

func (frh *FriendshipRequestHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		requestType, err := server.GetRequiredPathParam[string](ctx, appconstant.PathFriendRequestType)
		if err != nil {
			return nil, err
		}

		var response []dto.FriendshipRequestResponse
		switch requestType {
		case appconstant.SentFriendRequest:
			response, err = frh.svc.GetAllSent(ctx, userProfileID)
		case appconstant.ReceivedFriendRequest:
			response, err = frh.svc.GetAllReceived(ctx, userProfileID)
		default:
			err = ungerr.BadRequestError("invalid path parameter")
		}

		return response, err
	})
}

func (frh *FriendshipRequestHandler) HandleCancel() gin.HandlerFunc {
	return server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Cancel(ctx, userProfileID, requestID)
	})
}

func (frh *FriendshipRequestHandler) HandleIgnore() gin.HandlerFunc {
	return server.Handler(http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Ignore(ctx, userProfileID, requestID)
	})
}

func (frh *FriendshipRequestHandler) HandleBlock() gin.HandlerFunc {
	return server.Handler(http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		command := ctx.Query("command")
		switch command {
		case "block":
			return nil, frh.svc.Block(ctx, userProfileID, requestID)
		case "unblock":
			return nil, frh.svc.Unblock(ctx, userProfileID, requestID)
		default:
			return nil, ungerr.BadRequestError(fmt.Sprintf("unknown command: %s", command))
		}
	})
}

func (frh *FriendshipRequestHandler) HandleAccept() gin.HandlerFunc {
	return server.Handler(http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return frh.svc.Accept(ctx, userProfileID, requestID)
	})
}

func getIDs(ctx *gin.Context) (uuid.UUID, uuid.UUID, error) {
	userProfileID, err := getProfileID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	requestID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextFriendRequestID.String())
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userProfileID, requestID, nil
}
