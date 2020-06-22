package types

import (
	"strings"

	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"

	"github.com/tendermint/tendermint/crypto/merkle"
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
func NewQueryChannelResponse(channel Channel, proof *merkle.Proof, height int64) QueryChannelResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ChannelPath(portID, channelID), "/"))
	return QueryChannelResponse{
		Channel:     &channel,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   &path,
		ProofHeight: uint64(height),
	}
}

// NewQueryPacketResponse creates a new QueryPacketResponse instance
func NewQueryPacketResponse(
	portID, channelID string, sequence uint64, packet Packet, proof *merkle.Proof, height int64,
) QueryPacketResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/"))
	return QueryPacketResponse{
		Packet:      packet,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   &path,
		ProofHeight: uint64(height),
	}
}
