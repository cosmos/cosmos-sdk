package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ types.QueryServer = Keeper{}

// Channel implements the Query/Channel gRPC method
func (q Keeper) Channel(c context.Context, req *types.QueryChannelRequest) (*types.QueryChannelResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if err := host.PortIdentifierValidator(req.PortID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := host.ChannelIdentifierValidator(req.ChannelID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)
	channel, found := q.GetChannel(ctx, req.PortID, req.ChannelID)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrChannelNotFound, "port-id: , channel-id %s", req.PortID, req.ChannelID).Error(),
		)
	}

	return &types.QueryChannelResponse{
		Channel:     &channel,
		ProofHeight: uint64(ctx.BlockHeight()),
	}, nil
}

// Channels implements the Query/Channels gRPC method
func (q Keeper) Channels(c context.Context, req *types.QueryChannelsRequest) (*types.QueryChannelsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	channels := []types.IdentifiedChannel{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), host.KeyChannelPrefix)

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		var result types.IdentifiedChannel
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		channels = append(channels, result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryChannelsResponse{
		Channels: connections,
		Res:      res,
		Height:   ctx.BlockHeight(),
	}, nil
}
