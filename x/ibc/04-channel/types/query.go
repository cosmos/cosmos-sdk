package types

import (
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewQueryChannelResponse creates a new QueryChannelResponse instance
func NewQueryChannelResponse(portID, channelID string, channel Channel, proof []byte, height *clienttypes.Height) *QueryChannelResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ChannelPath(portID, channelID), "/"))
	return &QueryChannelResponse{
		Channel:     &channel,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: height,
	}
}

// NewQueryChannelClientStateResponse creates a newQueryChannelClientStateResponse instance
func NewQueryChannelClientStateResponse(identifiedClientState clienttypes.IdentifiedClientState, proof []byte, height *clienttypes.Height) *QueryChannelClientStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(identifiedClientState.ClientId, host.ClientStatePath()), "/"))
	return &QueryChannelClientStateResponse{
		IdentifiedClientState: &identifiedClientState,
		Proof:                 proof,
		ProofPath:             path.Pretty(),
		ProofHeight:           height,
	}
}

// NewQueryChannelConsensusStateResponse creates a newQueryChannelConsensusStateResponse instance
func NewQueryChannelConsensusStateResponse(clientID string, anyConsensusState *codectypes.Any, proof []byte, consensusStateHeight, height *clienttypes.Height) *QueryChannelConsensusStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(clientID, host.ConsensusStatePath(consensusStateHeight)), "/"))
	return &QueryChannelConsensusStateResponse{
		ConsensusState: anyConsensusState,
		ClientId:       clientID,
		Proof:          proof,
		ProofPath:      path.Pretty(),
		ProofHeight:    height,
	}
}

// NewQueryPacketCommitmentResponse creates a new QueryPacketCommitmentResponse instance
func NewQueryPacketCommitmentResponse(
	portID, channelID string, sequence uint64, commitment []byte, proof []byte, height *clienttypes.Height,
) *QueryPacketCommitmentResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketCommitmentPath(portID, channelID, sequence), "/"))
	return &QueryPacketCommitmentResponse{
		Commitment:  commitment,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: height,
	}
}

// NewQueryPacketAcknowledgementResponse creates a new QueryPacketAcknowledgementResponse instance
func NewQueryPacketAcknowledgementResponse(
	portID, channelID string, sequence uint64, acknowledgement []byte, proof []byte, height *clienttypes.Height,
) *QueryPacketAcknowledgementResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.PacketAcknowledgementPath(portID, channelID, sequence), "/"))
	return &QueryPacketAcknowledgementResponse{
		Acknowledgement: acknowledgement,
		Proof:           proof,
		ProofPath:       path.Pretty(),
		ProofHeight:     height,
	}
}

// NewQueryNextSequenceReceiveResponse creates a new QueryNextSequenceReceiveResponse instance
func NewQueryNextSequenceReceiveResponse(
	portID, channelID string, sequence uint64, proof []byte, height *clienttypes.Height,
) *QueryNextSequenceReceiveResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.NextSequenceRecvPath(portID, channelID), "/"))
	return &QueryNextSequenceReceiveResponse{
		NextSequenceReceive: sequence,
		Proof:               proof,
		ProofPath:           path.Pretty(),
		ProofHeight:         height,
	}
}
