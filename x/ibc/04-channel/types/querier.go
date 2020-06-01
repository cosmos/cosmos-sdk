package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// query routes supported by the IBC channel Querier
const (
	QueryAllChannels               = "channels"
	QueryChannel                   = "channel"
	QueryConnectionChannels        = "connection-channels"
	QueryPacketCommitments         = "packet-commitments"
	QueryUnrelayedAcknowledgements = "unrelayed-acknowledgements"
	QueryUnrelayedPacketSends      = "unrelayed-packet-sends"
)

// ChannelResponse defines the client query response for a channel which also
// includes a proof,its path and the height from which the proof was retrieved.
type ChannelResponse struct {
	Channel     IdentifiedChannel           `json:"channel" yaml:"channel"`
	Proof       commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewChannelResponse creates a new ChannelResponse instance
func NewChannelResponse(
	portID, channelID string, channel Channel, proof *merkle.Proof, height int64,
) ChannelResponse {
	return ChannelResponse{

		Channel:     NewIdentifiedChannel(portID, channelID, channel),
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   commitmenttypes.NewMerklePath(strings.Split(host.ChannelPath(portID, channelID), "/")),
		ProofHeight: uint64(height),
	}
}

// QueryAllChannelsParams defines the parameters necessary for querying for all
// channels.
type QueryAllChannelsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAllChannelsParams creates a new QueryAllChannelsParams instance.
func NewQueryAllChannelsParams(page, limit int) QueryAllChannelsParams {
	return QueryAllChannelsParams{
		Page:  page,
		Limit: limit,
	}
}

// QueryConnectionChannelsParams defines the parameters necessary for querying
// for all channels associated with a given connection.
type QueryConnectionChannelsParams struct {
	Connection string `json:"connection" yaml:"connection"`
	Page       int    `json:"page" yaml:"page"`
	Limit      int    `json:"limit" yaml:"limit"`
}

// NewQueryConnectionChannelsParams creates a new QueryConnectionChannelsParams instance.
func NewQueryConnectionChannelsParams(connection string, page, limit int) QueryConnectionChannelsParams {
	return QueryConnectionChannelsParams{
		Connection: connection,
		Page:       page,
		Limit:      limit,
	}
}

// QueryPacketCommitmentsParams defines the parameters necessary for querying
// all packet commitments at an associated port ID and channel ID.
type QueryPacketCommitmentsParams struct {
	PortID    string `json:"port_id" yaml:"port_id"`
	ChannelID string `json:"channel_id" yaml:"channel_id"`
	Page      int    `json:"page" yaml:"page"`
	Limit     int    `json:"limit" yaml:"limit"`
}

// NewQueryPacketCommitmentsParams creates a new QueryPacketCommitmentsParams instance.
func NewQueryPacketCommitmentsParams(portID, channelID string, page, limit int) QueryPacketCommitmentsParams {
	return QueryPacketCommitmentsParams{
		PortID:    portID,
		ChannelID: channelID,
		Page:      page,
		Limit:     limit,
	}
}

// QueryUnrelayedPacketsParams defines the parameters necessary for querying
// unrelayed packets at an associated port ID and channel ID.
type QueryUnrelayedPacketsParams struct {
	PortID    string   `json:"port_id" yaml:"port_id"`
	ChannelID string   `json:"channel_id" yaml:"channel_id"`
	Sequences []uint64 `json:"sequences" yaml:"sequences"`
	Page      int      `json:"page" yaml:"page"`
	Limit     int      `json:"limit" yaml:"limit"`
}

// NewQueryUnrealyedPacketsParams creates a new QueryUnrelayedPacketsParams instance.
func NewQueryUnrelayedPacketsParams(portID, channelID string, sequences []uint64, page, limit int) QueryUnrelayedPacketsParams {
	return QueryUnrelayedPacketsParams{
		PortID:    portID,
		ChannelID: channelID,
		Sequences: sequences,
		Page:      page,
		Limit:     limit,
	}
}

// PacketResponse defines the client query response for a packet which also
// includes a proof, its path and the height form which the proof was retrieved
type PacketResponse struct {
	Packet      Packet                      `json:"packet" yaml:"packet"`
	Proof       commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewPacketResponse creates a new PacketResponswe instance
func NewPacketResponse(
	portID, channelID string, sequence uint64, packet Packet, proof *merkle.Proof, height int64,
) PacketResponse {
	return PacketResponse{
		Packet:      packet,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/")),
		ProofHeight: uint64(height),
	}
}

// RecvResponse defines the client query response for the next receive sequence
// number which also includes a proof, its path and the height form which the
// proof was retrieved
type RecvResponse struct {
	NextSequenceRecv uint64                      `json:"next_sequence_recv" yaml:"next_sequence_recv"`
	Proof            commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath        commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight      uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewRecvResponse creates a new RecvResponse instance
func NewRecvResponse(
	portID, channelID string, sequenceRecv uint64, proof *merkle.Proof, height int64,
) RecvResponse {
	return RecvResponse{
		NextSequenceRecv: sequenceRecv,
		Proof:            commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:        commitmenttypes.NewMerklePath(strings.Split(host.NextSequenceRecvPath(portID, channelID), "/")),
		ProofHeight:      uint64(height),
	}
}
