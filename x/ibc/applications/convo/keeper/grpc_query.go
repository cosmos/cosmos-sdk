package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/applications/convo/types"
)

var _ types.QueryServer = Keeper{}

// PendingMessage implements the Query/PendingMessage gRPC method
func (q Keeper) PendingMessage(c context.Context, req *types.QueryPendingMessageRequest) (*types.QueryPendingMessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	msg := q.GetPendingMessage(ctx, req.Sender, req.Channel, req.Receiver)
	if msg == "" {
		return nil, sdkerrors.Wrapf(types.ErrNoMessageFound, "no pending message from %s to %s found over %s",
			req.Sender, req.Receiver, req.Channel)
	}

	return &types.QueryPendingMessageResponse{
		Message: msg,
	}, nil
}

// InboxMessage implements the Query/InboxMessage gRPC method
func (q Keeper) InboxMessage(c context.Context, req *types.QueryInboxMessageRequest) (*types.QueryInboxMessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	msg := q.GetInboxMessage(ctx, req.Sender, req.Channel, req.Receiver)
	if msg == "" {
		return nil, sdkerrors.Wrapf(types.ErrNoMessageFound, "no inbox message from %s to %s found over %s",
			req.Sender, req.Receiver, req.Channel)
	}

	return &types.QueryInboxMessageResponse{
		Message: msg,
	}, nil
}

// OutboxMessage implements the Query/OutboxMessage gRPC method
func (q Keeper) OutboxMessage(c context.Context, req *types.QueryOutboxMessageRequest) (*types.QueryOutboxMessageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	msg := q.GetOutboxMessage(ctx, req.Sender, req.Channel, req.Receiver)
	if msg == "" {
		return nil, sdkerrors.Wrapf(types.ErrNoMessageFound, "no inbox message from %s to %s found over %s",
			req.Sender, req.Receiver, req.Channel)
	}

	return &types.QueryOutboxMessageResponse{
		Message: msg,
	}, nil
}
