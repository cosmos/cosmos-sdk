package keeper

import (
	"context"

	connectiontypes "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/types"
)

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

// PacketCommitment implements the IBC QueryServer interface
func (q Keeper) PacketCommitment(c context.Context, req *channeltypes.QueryPacketCommitmentRequest) (*channeltypes.QueryPacketCommitmentResponse, error) {
	return q.ChannelKeeper.PacketCommitment(c, req)
}

// PacketCommitments implements the IBC QueryServer interface
func (q Keeper) PacketCommitments(c context.Context, req *channeltypes.QueryPacketCommitmentsRequest) (*channeltypes.QueryPacketCommitmentsResponse, error) {
	return q.ChannelKeeper.PacketCommitments(c, req)
}

// PacketAcknowledgement implements the IBC QueryServer interface
func (q Keeper) PacketAcknowledgement(c context.Context, req *channeltypes.QueryPacketAcknowledgementRequest) (*channeltypes.QueryPacketAcknowledgementResponse, error) {
	return q.ChannelKeeper.PacketAcknowledgement(c, req)
}

// UnrelayedPackets implements the IBC QueryServer interface
func (q Keeper) UnrelayedPackets(c context.Context, req *channeltypes.QueryUnrelayedPacketsRequest) (*channeltypes.QueryUnrelayedPacketsResponse, error) {
	return q.ChannelKeeper.UnrelayedPackets(c, req)
}

// NextSequenceReceive implements the IBC QueryServer interface
func (q Keeper) NextSequenceReceive(c context.Context, req *channeltypes.QueryNextSequenceReceiveRequest) (*channeltypes.QueryNextSequenceReceiveResponse, error) {
	return q.ChannelKeeper.NextSequenceReceive(c, req)
}
