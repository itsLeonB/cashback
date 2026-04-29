package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/itsLeonB/cashback/internal/appconstant"
	"github.com/itsLeonB/cashback/internal/domain/dto"
	"github.com/itsLeonB/cashback/internal/domain/service"
	_ "github.com/itsLeonB/ginkgo/pkg/response"
	"github.com/itsLeonB/ginkgo/pkg/server"
	"github.com/itsLeonB/ungerr"
)

type FriendshipRequestHandler struct {
	svc service.FriendshipRequestService
}

func NewFriendshipRequestHandler(svc service.FriendshipRequestService) *FriendshipRequestHandler {
	return &FriendshipRequestHandler{svc}
}

// HandleSend godoc
// @Summary      Send a friend request
// @Tags         friend-requests
// @Security     BearerAuth
// @Produce      json
// @Param        profileId path string true "Target profile ID"
// @Success      201  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /profiles/{profileId}/friend-requests [post]
func (frh *FriendshipRequestHandler) HandleSend() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleSend", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, err := getProfileID(ctx)
		if err != nil {
			return nil, err
		}

		friendProfileID, err := server.GetRequiredPathParam[uuid.UUID](ctx, appconstant.ContextProfileID.String())
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Send(ctx.Request.Context(), userProfileID, friendProfileID)
	})
}

// HandleGetAll godoc
// @Summary      Get all friend requests (sent or received)
// @Tags         friend-requests
// @Security     BearerAuth
// @Produce      json
// @Param        type path string true "Request type: sent or received"
// @Success      200  {object}  response.JSONResponse[[]dto.FriendshipRequestResponse]
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /friend-requests/{type} [get]
func (frh *FriendshipRequestHandler) HandleGetAll() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleGetAll", http.StatusOK, func(ctx *gin.Context) (any, error) {
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
			response, err = frh.svc.GetAllSent(ctx.Request.Context(), userProfileID)
		case appconstant.ReceivedFriendRequest:
			response, err = frh.svc.GetAllReceived(ctx.Request.Context(), userProfileID)
		default:
			err = ungerr.BadRequestError("invalid path parameter")
		}

		return response, err
	})
}

// HandleCancel godoc
// @Summary      Cancel a sent friend request
// @Tags         friend-requests
// @Security     BearerAuth
// @Param        requestId path string true "Friend request ID"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /friend-requests/sent/{requestId} [delete]
func (frh *FriendshipRequestHandler) HandleCancel() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleCancel", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Cancel(ctx.Request.Context(), userProfileID, requestID)
	})
}

// HandleIgnore godoc
// @Summary      Ignore a received friend request
// @Tags         friend-requests
// @Security     BearerAuth
// @Param        requestId path string true "Friend request ID"
// @Success      204
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /friend-requests/received/{requestId} [delete]
func (frh *FriendshipRequestHandler) HandleIgnore() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleIgnore", http.StatusNoContent, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return nil, frh.svc.Ignore(ctx.Request.Context(), userProfileID, requestID)
	})
}

// HandleBlock godoc
// @Summary      Block or unblock a friend request sender
// @Tags         friend-requests
// @Security     BearerAuth
// @Produce      json
// @Param        requestId path  string true  "Friend request ID"
// @Param        command   query string true  "Command: block or unblock"
// @Success      200  {object}  map[string]any
// @Failure      400  {object}  map[string]any
// @Failure      401  {object}  map[string]any
// @Router       /friend-requests/received/{requestId} [patch]
func (frh *FriendshipRequestHandler) HandleBlock() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleBlock", http.StatusOK, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		command := ctx.Query("command")
		switch command {
		case "block":
			return nil, frh.svc.Block(ctx.Request.Context(), userProfileID, requestID)
		case "unblock":
			return nil, frh.svc.Unblock(ctx.Request.Context(), userProfileID, requestID)
		default:
			return nil, ungerr.BadRequestError(fmt.Sprintf("unknown command: %s", command))
		}
	})
}

// HandleAccept godoc
// @Summary      Accept a received friend request
// @Tags         friend-requests
// @Security     BearerAuth
// @Produce      json
// @Param        requestId path string true "Friend request ID"
// @Success      201  {object}  response.JSONResponse[dto.FriendshipResponse]
// @Failure      401  {object}  map[string]any
// @Failure      404  {object}  map[string]any
// @Router       /friend-requests/received/{requestId} [post]
func (frh *FriendshipRequestHandler) HandleAccept() gin.HandlerFunc {
	return server.Handler("FriendshipRequestHandler.HandleAccept", http.StatusCreated, func(ctx *gin.Context) (any, error) {
		userProfileID, requestID, err := getIDs(ctx)
		if err != nil {
			return nil, err
		}

		return frh.svc.Accept(ctx.Request.Context(), userProfileID, requestID)
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
