package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/ginkgo"
	"github.com/itsLeonB/orcashtrator/internal/appconstant"
	"github.com/itsLeonB/orcashtrator/internal/dto"
	"github.com/itsLeonB/orcashtrator/internal/service"
	"github.com/itsLeonB/orcashtrator/internal/util"
	"github.com/itsLeonB/ungerr"
)

type FriendshipRequestHandler struct {
	svc service.FriendshipRequestService
}

func NewFriendshipRequestHandler(svc service.FriendshipRequestService) *FriendshipRequestHandler {
	return &FriendshipRequestHandler{svc}
}

func (frh *FriendshipRequestHandler) HandleSend() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		friendProfileID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return 0, "", nil, err
		}
		if err = frh.svc.Send(ctx, userProfileID, friendProfileID); err != nil {
			return 0, "", nil, err
		}
		return http.StatusCreated, appconstant.MsgInsertData, nil, nil
	})
}

func (frh *FriendshipRequestHandler) HandleGetAll() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, err := util.GetProfileID(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		requestType, err := ginkgo.GetRequiredPathParam[string](ctx, appconstant.PathFriendRequestType)
		if err != nil {
			return 0, "", nil, err
		}

		var response []dto.FriendshipRequestResponse
		switch requestType {
		case appconstant.SentFriendRequest:
			response, err = frh.svc.GetAllSent(ctx, userProfileID)
		case appconstant.ReceivedFriendRequest:
			response, err = frh.svc.GetAllReceived(ctx, userProfileID)
		default:
			return 0, "", nil, ungerr.BadRequestError("invalid path parameter")
		}
		if err != nil {
			return 0, "", nil, err
		}

		return http.StatusOK, appconstant.MsgGetData, response, nil
	})
}

func (frh *FriendshipRequestHandler) HandleCancel() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		if err = frh.svc.Cancel(ctx, userProfileID, requestID); err != nil {
			return 0, "", nil, err
		}
		return http.StatusNoContent, "", nil, nil
	})
}

func (frh *FriendshipRequestHandler) HandleIgnore() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		if err = frh.svc.Ignore(ctx, userProfileID, requestID); err != nil {
			return 0, "", nil, err
		}
		return http.StatusNoContent, "", nil, nil
	})
}

func (frh *FriendshipRequestHandler) HandleBlock() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return 0, "", nil, err
		}

		command := ctx.Query("command")
		switch command {
		case "block":
			err = frh.svc.Block(ctx, userProfileID, requestID)
		case "unblock":
			err = frh.svc.Unblock(ctx, userProfileID, requestID)
		default:
			return 0, "", nil, ungerr.BadRequestError(fmt.Sprintf("unknown command: %s", command))
		}
		if err != nil {
			return 0, "", nil, err
		}

		return http.StatusOK, appconstant.MsgUpdateData, nil, nil
	})
}

func (frh *FriendshipRequestHandler) HandleAccept() gin.HandlerFunc {
	return ginkgo.WrapHandler(func(ctx *gin.Context) (int, string, any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return 0, "", nil, err
		}
		response, err := frh.svc.Accept(ctx, userProfileID, requestID)
		if err != nil {
			return 0, "", nil, err
		}
		return http.StatusCreated, appconstant.MsgInsertData, response, nil
	})
}

func getIDs(ctx *gin.Context) (uuid.UUID, uuid.UUID, error) {
	userProfileID, err := util.GetProfileID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	requestID, err := ginkgo.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextFriendRequestID.String())
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return userProfileID, requestID, nil
}
