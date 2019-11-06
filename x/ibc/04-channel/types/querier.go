package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// query routes supported by the IBC channel Querier
const (
	QueryChannel = "channel"
)

// ChannelResponse defines the client query response for a channel which also
// includes a proof,its path and the height from which the proof was retrieved.
type ChannelResponse struct {
	Channel     Channel          `json:"channel" yaml:"channel"`
	Proof       commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewChannelResponse creates a new ChannelResponse instance
func NewChannelResponse(
	portID, channelID string, channel Channel, proof *merkle.Proof, height int64,
) ChannelResponse {
	return ChannelResponse{
		Channel:     channel,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(ChannelPath(portID, channelID), "/")),
		ProofHeight: uint64(height),
	}
}

// QueryChannelParams defines the params for the following queries:
// - 'custom/ibc/channel'
type QueryChannelParams struct {
	PortID    string
	ChannelID string
}

// NewQueryChannelParams creates a new QueryChannelParams instance
func NewQueryChannelParams(portID, channelID string) QueryChannelParams {
	return QueryChannelParams{
		PortID:    portID,
		ChannelID: channelID,
	}
}

// PacketResponse defines the client query response for a packet which also
// includes a proof, its path and the height form which the proof was retrieved
type PacketResponse struct {
	Packet      Packet           `json:"packet" yaml:"packet"`
	Proof       commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewPacketResponse creates a new PacketResponswe instance
func NewPacketResponse(
	portID, channelID string, sequence uint64, packet Packet, proof *merkle.Proof, height int64,
) PacketResponse {
	return PacketResponse{
		Packet:      packet,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(PacketCommitmentPath(portID, channelID, sequence), "/")),
		ProofHeight: uint64(height),
	}
}
