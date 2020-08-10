package types

import (
	"strings"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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

// NewQueryChannelClientStateResponse creates a newQueryChannelClientStateResponse instance
func NewQueryChannelClientStateResponse(identifiedClientState clienttypes.IdentifiedClientState, proof []byte, height int64) *QueryChannelClientStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(identifiedClientState.ID, host.ClientStatePath()), "/"))
	return &QueryChannelClientStateResponse{
		IdentifiedClientState: &identifiedClientState,
		Proof:                 proof,
		ProofPath:             path.Pretty(),
		ProofHeight:           uint64(height),
	}
}

// NewQueryChannelConsensusStateResponse creates a newQueryChannelConsensusStateResponse instance
func NewQueryChannelConsensusStateResponse(clientID string, consensusState clientexported.ConsensusState, proof []byte, height int64) *QueryChannelConsensusStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(clientID, host.ConsensusStatePath(consensusState.GetHeight())), "/"))
	anyConsensusState := clienttypes.GetAnyFromConsensusState(consensusState)
	return &QueryChannelConsensusStateResponse{
		ConsensusState: anyConsensusState,
		ClientID:       clientID,
		Proof:          proof,
		ProofPath:      path.Pretty(),
		ProofHeight:    uint64(height),
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

// NewQueryPacketAcknowledgementResponse creates a new QueryPacketAcknowledgementResponse instance
func NewQueryPacketAcknowledgementResponse(
	portID, channelID string, sequence uint64, acknowledgement []byte, proof []byte, height int64,
) *QueryPacketAcknowledgementResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketAcknowledgementPath(portID, channelID, sequence), "/"))
	return &QueryPacketAcknowledgementResponse{
		Acknowledgement: acknowledgement,
		Proof:           proof,
		ProofPath:       path.Pretty(),
		ProofHeight:     uint64(height),
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
