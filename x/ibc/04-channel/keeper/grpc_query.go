package keeper

import (
	"context"
	"strconv"
	"strings"

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

	return types.NewQueryChannelResponse(req.PortID, req.ChannelID, channel, nil, ctx.BlockHeight()), nil
}

// Channels implements the Query/Channels gRPC method
func (q Keeper) Channels(c context.Context, req *types.QueryChannelsRequest) (*types.QueryChannelsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	channels := []*types.IdentifiedChannel{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.KeyChannelPrefix))

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		var result types.IdentifiedChannel
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		channels = append(channels, &result)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryChannelsResponse{
		Channels: channels,
		Res:      res,
		Height:   ctx.BlockHeight(),
	}, nil
}

// ConnectionChannels implements the Query/ConnectionChannels gRPC method
func (q Keeper) ConnectionChannels(c context.Context, req *types.QueryConnectionChannelsRequest) (*types.QueryConnectionChannelsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	channels := []*types.IdentifiedChannel{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.KeyChannelPrefix))

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		var channel types.IdentifiedChannel
		if err := q.cdc.UnmarshalBinaryBare(value, &channel); err != nil {
			return err
		}

		// ignore channel and continue to the next item if the connection is
		// different than the requested one
		if channel.ConnectionHops[0] != req.Connection {
			return nil
		}

		channels = append(channels, &channel)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryConnectionChannelsResponse{
		Channels: channels,
		Res:      res,
		Height:   ctx.BlockHeight(),
	}, nil
}

// PacketCommitment implements the Query/PacketCommitment gRPC method
func (q Keeper) PacketCommitment(c context.Context, req *types.QueryPacketCommitmentRequest) (*types.QueryPacketCommitmentResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	if err := host.PortIdentifierValidator(req.PortID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := host.ChannelIdentifierValidator(req.ChannelID); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if req.Sequence == 0 {
		return nil, status.Error(codes.InvalidArgument, "packet sequence cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	commitmentBz := q.GetPacketCommitment(ctx, req.PortID, req.ChannelID, req.Sequence)
	if len(commitmentBz) == 0 {
		return nil, status.Error(codes.NotFound, "packet commitment hash not found")
	}

	return types.NewQueryPacketCommitmentResponse(req.PortID, req.ChannelID, req.Sequence, commitmentBz, nil, ctx.BlockHeight()), nil
}

// PacketCommitments implements the Query/PacketCommitments gRPC method
func (q Keeper) PacketCommitments(c context.Context, req *types.QueryPacketCommitmentsRequest) (*types.QueryPacketCommitmentsResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	commitments := []*types.PacketAckCommitment{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.PacketCommitmentPrefixPath(req.PortID, req.ChannelID)))

	res, err := query.Paginate(store, req.Req, func(key []byte, value []byte) error {
		keySplit := strings.Split(string(key), "/")

		sequence, err := strconv.ParseUint(keySplit[len(keySplit)-1], 10, 64)
		if err != nil {
			return err
		}

		commitment := types.NewPacketAckCommitment(req.PortID, req.ChannelID, sequence, value)
		commitments = append(commitments, &commitment)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &types.QueryPacketCommitmentsResponse{
		Commitments: commitments,
		Res:         res,
		Height:      ctx.BlockHeight(),
	}, nil
}
