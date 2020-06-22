package types

import (
	"strings"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// query routes supported by the IBC channel Querier
const (
	QueryAllChannels               = "channels"
	QueryChannel                   = "channel"
	QueryConnectionChannels        = "connection-channels"
	QueryChannelClientState        = "channel-client-state"
	QueryPacketCommitments         = "packet-commitments"
	QueryUnrelayedAcknowledgements = "unrelayed-acknowledgements"
	QueryUnrelayedPacketSends      = "unrelayed-packet-sends"
)

// NewQueryChannelResponse creates a new QueryChannelResponse instance
func NewQueryChannelResponse(portID, channelID string, channel Channel, proof []byte, height int64) *QueryChannelResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ChannelPath(portID, channelID), "/"))
	return &QueryChannelResponse{
		Channel:     &channel,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: uint64(height),
	}
}

// NewQueryPacketResponse creates a new QueryPacketResponse instance
func NewQueryPacketResponse(
	portID, channelID string, sequence uint64, packet Packet, proof []byte, height int64,
) *QueryPacketResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/"))
	return &QueryPacketResponse{
		Packet:      &packet,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: uint64(height),
	}
}

// NewQueryChannelClientStateRequest creates a new QueryChannelClientStateRequest instance.
func NewQueryChannelClientStateRequest(portID, channelID string) *QueryChannelClientStateRequest {
	return &QueryChannelClientStateRequest{
		PortID:    portID,
		ChannelID: channelID,
	}
}
