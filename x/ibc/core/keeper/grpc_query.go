package keeper

import (
	"context"

	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/core/03-connection/types"
	channeltypes "github.com/cosmos/cosmos-sdk/x/ibc/core/04-channel/types"
)

// ClientState implements the IBC QueryServer interface
func (q Keeper) ClientState(c context.Context, req *clienttypes.QueryClientStateRequest) (*clienttypes.QueryClientStateResponse, error) {
	return q.ClientKeeper.ClientState(c, req)
}

// ClientStates implements the IBC QueryServer interface
func (q Keeper) ClientStates(c context.Context, req *clienttypes.QueryClientStatesRequest) (*clienttypes.QueryClientStatesResponse, error) {
	return q.ClientKeeper.ClientStates(c, req)
}

// ConsensusState implements the IBC QueryServer interface
func (q Keeper) ConsensusState(c context.Context, req *clienttypes.QueryConsensusStateRequest) (*clienttypes.QueryConsensusStateResponse, error) {
	return q.ClientKeeper.ConsensusState(c, req)
}

// ConsensusStates implements the IBC QueryServer interface
func (q Keeper) ConsensusStates(c context.Context, req *clienttypes.QueryConsensusStatesRequest) (*clienttypes.QueryConsensusStatesResponse, error) {
	return q.ClientKeeper.ConsensusStates(c, req)
}

// ClientParams implements the IBC QueryServer interface
func (q Keeper) ClientParams(c context.Context, req *clienttypes.QueryClientParamsRequest) (*clienttypes.QueryClientParamsResponse, error) {
	return q.ClientKeeper.ClientParams(c, req)
}

// Connection implements the IBC QueryServer interface
func (q Keeper) Connection(c context.Context, req *connectiontypes.QueryConnectionRequest) (*connectiontypes.QueryConnectionResponse, error) {
	return q.ConnectionKeeper.Connection(c, req)
}

// Connections implements the IBC QueryServer interface
func (q Keeper) Connections(c context.Context, req *connectiontypes.QueryConnectionsRequest) (*connectiontypes.QueryConnectionsResponse, error) {
	return q.ConnectionKeeper.Connections(c, req)
}

// ClientConnections implements the IBC QueryServer interface
func (q Keeper) ClientConnections(c context.Context, req *connectiontypes.QueryClientConnectionsRequest) (*connectiontypes.QueryClientConnectionsResponse, error) {
	return q.ConnectionKeeper.ClientConnections(c, req)
}

// ConnectionClientState implements the IBC QueryServer interface
func (q Keeper) ConnectionClientState(c context.Context, req *connectiontypes.QueryConnectionClientStateRequest) (*connectiontypes.QueryConnectionClientStateResponse, error) {
	return q.ConnectionKeeper.ConnectionClientState(c, req)
}

// ConnectionConsensusState implements the IBC QueryServer interface
func (q Keeper) ConnectionConsensusState(c context.Context, req *connectiontypes.QueryConnectionConsensusStateRequest) (*connectiontypes.QueryConnectionConsensusStateResponse, error) {
	return q.ConnectionKeeper.ConnectionConsensusState(c, req)
}

// Channel implements the IBC QueryServer interface
func (q Keeper) Channel(c context.Context, req *channeltypes.QueryChannelRequest) (*channeltypes.QueryChannelResponse, error) {
	return q.ChannelKeeper.Channel(c, req)
}

// Channels implements the IBC QueryServer interface
func (q Keeper) Channels(c context.Context, req *channeltypes.QueryChannelsRequest) (*channeltypes.QueryChannelsResponse, error) {
	return q.ChannelKeeper.Channels(c, req)
}

// ConnectionChannels implements the IBC QueryServer interface
func (q Keeper) ConnectionChannels(c context.Context, req *channeltypes.QueryConnectionChannelsRequest) (*channeltypes.QueryConnectionChannelsResponse, error) {
	return q.ChannelKeeper.ConnectionChannels(c, req)
}

// ChannelClientState implements the IBC QueryServer interface
func (q Keeper) ChannelClientState(c context.Context, req *channeltypes.QueryChannelClientStateRequest) (*channeltypes.QueryChannelClientStateResponse, error) {
	return q.ChannelKeeper.ChannelClientState(c, req)
}

// ChannelConsensusState implements the IBC QueryServer interface
func (q Keeper) ChannelConsensusState(c context.Context, req *channeltypes.QueryChannelConsensusStateRequest) (*channeltypes.QueryChannelConsensusStateResponse, error) {
	return q.ChannelKeeper.ChannelConsensusState(c, req)
}

// PacketCommitment implements the IBC QueryServer interface
func (q Keeper) PacketCommitment(c context.Context, req *channeltypes.QueryPacketCommitmentRequest) (*channeltypes.QueryPacketCommitmentResponse, error) {
	return q.ChannelKeeper.PacketCommitment(c, req)
}

// PacketCommitments implements the IBC QueryServer interface
func (q Keeper) PacketCommitments(c context.Context, req *channeltypes.QueryPacketCommitmentsRequest) (*channeltypes.QueryPacketCommitmentsResponse, error) {
	return q.ChannelKeeper.PacketCommitments(c, req)
}

// PacketReceipt implements the IBC QueryServer interface
func (q Keeper) PacketReceipt(c context.Context, req *channeltypes.QueryPacketReceiptRequest) (*channeltypes.QueryPacketReceiptResponse, error) {
	return q.ChannelKeeper.PacketReceipt(c, req)
}

// PacketAcknowledgement implements the IBC QueryServer interface
func (q Keeper) PacketAcknowledgement(c context.Context, req *channeltypes.QueryPacketAcknowledgementRequest) (*channeltypes.QueryPacketAcknowledgementResponse, error) {
	return q.ChannelKeeper.PacketAcknowledgement(c, req)
}

// PacketAcknowledgements implements the IBC QueryServer interface
func (q Keeper) PacketAcknowledgements(c context.Context, req *channeltypes.QueryPacketAcknowledgementsRequest) (*channeltypes.QueryPacketAcknowledgementsResponse, error) {
	return q.ChannelKeeper.PacketAcknowledgements(c, req)
}

// UnreceivedPackets implements the IBC QueryServer interface
func (q Keeper) UnreceivedPackets(c context.Context, req *channeltypes.QueryUnreceivedPacketsRequest) (*channeltypes.QueryUnreceivedPacketsResponse, error) {
	return q.ChannelKeeper.UnreceivedPackets(c, req)
}

// UnreceivedAcks implements the IBC QueryServer interface
func (q Keeper) UnreceivedAcks(c context.Context, req *channeltypes.QueryUnreceivedAcksRequest) (*channeltypes.QueryUnreceivedAcksResponse, error) {
	return q.ChannelKeeper.UnreceivedAcks(c, req)
}

// NextSequenceReceive implements the IBC QueryServer interface
func (q Keeper) NextSequenceReceive(c context.Context, req *channeltypes.QueryNextSequenceReceiveRequest) (*channeltypes.QueryNextSequenceReceiveResponse, error) {
	return q.ChannelKeeper.NextSequenceReceive(c, req)
}
