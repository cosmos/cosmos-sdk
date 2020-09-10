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
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var _ types.QueryServer = (*Keeper)(nil)

// Channel implements the Query/Channel gRPC method
func (q Keeper) Channel(c context.Context, req *types.QueryChannelRequest) (*types.QueryChannelResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	channel, found := q.GetChannel(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrChannelNotFound, "port-id: %s, channel-id %s", req.PortId, req.ChannelId).Error(),
		)
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryChannelResponse(req.PortId, req.ChannelId, channel, nil, selfHeight), nil
}

// Channels implements the Query/Channels gRPC method
func (q Keeper) Channels(c context.Context, req *types.QueryChannelsRequest) (*types.QueryChannelsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)

	channels := []*types.IdentifiedChannel{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.KeyChannelPrefix))

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var result types.Channel
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		portID, channelID, err := host.ParseChannelPath(string(key))
		if err != nil {
			return err
		}

		identifiedChannel := types.NewIdentifiedChannel(portID, channelID, result)
		channels = append(channels, &identifiedChannel)
		return nil
	})

	if err != nil {
		return nil, err
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return &types.QueryChannelsResponse{
		Channels:   channels,
		Pagination: pageRes,
		Height:     selfHeight,
	}, nil
}

// ConnectionChannels implements the Query/ConnectionChannels gRPC method
func (q Keeper) ConnectionChannels(c context.Context, req *types.QueryConnectionChannelsRequest) (*types.QueryConnectionChannelsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := host.ConnectionIdentifierValidator(req.Connection); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	ctx := sdk.UnwrapSDKContext(c)

	channels := []*types.IdentifiedChannel{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.KeyChannelPrefix))

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		var result types.Channel
		if err := q.cdc.UnmarshalBinaryBare(value, &result); err != nil {
			return err
		}

		// ignore channel and continue to the next item if the connection is
		// different than the requested one
		if result.ConnectionHops[0] != req.Connection {
			return nil
		}

		portID, channelID, err := host.ParseChannelPath(string(key))
		if err != nil {
			return err
		}

		identifiedChannel := types.NewIdentifiedChannel(portID, channelID, result)
		channels = append(channels, &identifiedChannel)
		return nil
	})

	if err != nil {
		return nil, err
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return &types.QueryConnectionChannelsResponse{
		Channels:   channels,
		Pagination: pageRes,
		Height:     selfHeight,
	}, nil
}

// ChannelClientState implements the Query/ChannelClientState gRPC method
func (q Keeper) ChannelClientState(c context.Context, req *types.QueryChannelClientStateRequest) (*types.QueryChannelClientStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	channel, found := q.GetChannel(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrChannelNotFound, "port-id: %s, channel-id %s", req.PortId, req.ChannelId).Error(),
		)
	}

	connection, found := q.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", channel.ConnectionHops[0]).Error(),
		)
	}

	clientState, found := q.clientKeeper.GetClientState(ctx, connection.ClientId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(clienttypes.ErrClientNotFound, "client-id: %s", connection.ClientId).Error(),
		)
	}

	identifiedClientState := clienttypes.NewIdentifiedClientState(connection.ClientId, clientState)

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryChannelClientStateResponse(identifiedClientState, nil, selfHeight), nil
}

// ChannelConsensusState implements the Query/ChannelConsensusState gRPC method
func (q Keeper) ChannelConsensusState(c context.Context, req *types.QueryChannelConsensusStateRequest) (*types.QueryChannelConsensusStateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	channel, found := q.GetChannel(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrChannelNotFound, "port-id: %s, channel-id %s", req.PortId, req.ChannelId).Error(),
		)
	}

	connection, found := q.connectionKeeper.GetConnection(ctx, channel.ConnectionHops[0])
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(connectiontypes.ErrConnectionNotFound, "connection-id: %s", channel.ConnectionHops[0]).Error(),
		)
	}

	consHeight := clienttypes.NewHeight(req.EpochNumber, req.EpochHeight)
	consensusState, found := q.clientKeeper.GetClientConsensusState(ctx, connection.ClientId, consHeight)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(clienttypes.ErrConsensusStateNotFound, "client-id: %s", connection.ClientId).Error(),
		)
	}

	anyConsensusState, err := clienttypes.PackConsensusState(consensusState)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryChannelConsensusStateResponse(connection.ClientId, anyConsensusState, consHeight, nil, selfHeight), nil
}

// PacketCommitment implements the Query/PacketCommitment gRPC method
func (q Keeper) PacketCommitment(c context.Context, req *types.QueryPacketCommitmentRequest) (*types.QueryPacketCommitmentResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	if req.Sequence == 0 {
		return nil, status.Error(codes.InvalidArgument, "packet sequence cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	commitmentBz := q.GetPacketCommitment(ctx, req.PortId, req.ChannelId, req.Sequence)
	if len(commitmentBz) == 0 {
		return nil, status.Error(codes.NotFound, "packet commitment hash not found")
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryPacketCommitmentResponse(req.PortId, req.ChannelId, req.Sequence, commitmentBz, nil, selfHeight), nil
}

// PacketCommitments implements the Query/PacketCommitments gRPC method
func (q Keeper) PacketCommitments(c context.Context, req *types.QueryPacketCommitmentsRequest) (*types.QueryPacketCommitmentsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	commitments := []*types.PacketAckCommitment{}
	store := prefix.NewStore(ctx.KVStore(q.storeKey), []byte(host.PacketCommitmentPrefixPath(req.PortId, req.ChannelId)))

	pageRes, err := query.Paginate(store, req.Pagination, func(key, value []byte) error {
		keySplit := strings.Split(string(key), "/")

		sequence, err := strconv.ParseUint(keySplit[len(keySplit)-1], 10, 64)
		if err != nil {
			return err
		}

		commitment := types.NewPacketAckCommitment(req.PortId, req.ChannelId, sequence, value)
		commitments = append(commitments, &commitment)
		return nil
	})

	if err != nil {
		return nil, err
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return &types.QueryPacketCommitmentsResponse{
		Commitments: commitments,
		Pagination:  pageRes,
		Height:      selfHeight,
	}, nil
}

// PacketAcknowledgement implements the Query/PacketAcknowledgement gRPC method
func (q Keeper) PacketAcknowledgement(c context.Context, req *types.QueryPacketAcknowledgementRequest) (*types.QueryPacketAcknowledgementResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	if req.Sequence == 0 {
		return nil, status.Error(codes.InvalidArgument, "packet sequence cannot be 0")
	}

	ctx := sdk.UnwrapSDKContext(c)

	acknowledgementBz, found := q.GetPacketAcknowledgement(ctx, req.PortId, req.ChannelId, req.Sequence)
	if !found || len(acknowledgementBz) == 0 {
		return nil, status.Error(codes.NotFound, "packet acknowledgement hash not found")
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryPacketAcknowledgementResponse(req.PortId, req.ChannelId, req.Sequence, acknowledgementBz, nil, selfHeight), nil
}

// UnrelayedPackets implements the Query/UnrelayedPackets gRPC method. Given
// a list of counterparty packet commitments, the querier checks if the packet
// sequence has an acknowledgement stored. If req.Acknowledgements is true then
// all unrelayed acknowledgements are returned (ack exists), otherwise all
// unrelayed packet commitments are returned (ack does not exist).
//
// NOTE: The querier makes the assumption that the provided list of packet
// commitments is correct and will not function properly if the list
// is not up to date. Ideally the query height should equal the latest height
// on the counterparty's client which represents this chain.
func (q Keeper) UnrelayedPackets(c context.Context, req *types.QueryUnrelayedPacketsRequest) (*types.QueryUnrelayedPacketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)

	var unrelayedSequences = []uint64{}

	for i, seq := range req.PacketCommitmentSequences {
		if seq == 0 {
			return nil, status.Errorf(codes.InvalidArgument, "packet sequence %d cannot be 0", i)
		}

		// if req.Acknowledgements is true append sequences with an existing acknowledgement
		// otherwise append sequences without an existing acknowledgement.
		if _, found := q.GetPacketAcknowledgement(ctx, req.PortId, req.ChannelId, seq); found == req.Acknowledgements {
			unrelayedSequences = append(unrelayedSequences, seq)
		}

	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return &types.QueryUnrelayedPacketsResponse{
		Sequences: unrelayedSequences,
		Height:    selfHeight,
	}, nil
}

// NextSequenceReceive implements the Query/NextSequenceReceive gRPC method
func (q Keeper) NextSequenceReceive(c context.Context, req *types.QueryNextSequenceReceiveRequest) (*types.QueryNextSequenceReceiveResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if err := validategRPCRequest(req.PortId, req.ChannelId); err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(c)
	sequence, found := q.GetNextSequenceRecv(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(
			codes.NotFound,
			sdkerrors.Wrapf(types.ErrSequenceReceiveNotFound, "port-id: %s, channel-id %s", req.PortId, req.ChannelId).Error(),
		)
	}

	selfHeight := clienttypes.GetSelfHeight(ctx)
	return types.NewQueryNextSequenceReceiveResponse(req.PortId, req.ChannelId, sequence, nil, selfHeight), nil
}

func validategRPCRequest(portID, channelID string) error {
	if err := host.PortIdentifierValidator(portID); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := host.ChannelIdentifierValidator(channelID); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}
