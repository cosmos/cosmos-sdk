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
	QueryChannelConsensusState     = "channel-consensus-state"
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

// NewQueryPacketCommitmentResponse creates a new QueryPacketCommitmentResponse instance
func NewQueryPacketCommitmentResponse(
	portID, channelID string, sequence uint64, commitment []byte, proof []byte, height int64,
) *QueryPacketCommitmentResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/"))
	return &QueryPacketCommitmentResponse{
		Commitment:  commitment,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: uint64(height),
	}
}

// NewQueryNextSequenceReceiveResponse creates a new QueryNextSequenceReceiveResponse instance
func NewQueryNextSequenceReceiveResponse(
	portID, channelID string, sequence uint64, proof []byte, height int64,
) *QueryNextSequenceReceiveResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.NextSequenceRecvPath(portID, channelID), "/"))
	return &QueryNextSequenceReceiveResponse{
		NextSequenceReceive: sequence,
		Proof:               proof,
		ProofPath:           path.Pretty(),
		ProofHeight:         uint64(height),
	}
}

// NewQueryChannelClientStateRequest creates a new QueryChannelClientStateRequest instance.
func NewQueryChannelClientStateRequest(portID, channelID string) *QueryChannelClientStateRequest {
	return &QueryChannelClientStateRequest{
		PortID:    portID,
		ChannelID: channelID,
	}
}

// NewQueryChannelConsensusStateRequest creates a new QueryChannelConsensusStateRequest instance.
func NewQueryChannelConsensusStateRequest(portID, channelID string) *QueryChannelConsensusStateRequest {
	return &QueryChannelConsensusStateRequest{
		PortID:    portID,
		ChannelID: channelID,
	}
}
