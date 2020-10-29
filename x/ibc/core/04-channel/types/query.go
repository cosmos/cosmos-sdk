package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// NewQueryChannelResponse creates a new QueryChannelResponse instance
func NewQueryChannelResponse(channel Channel, proof []byte, height clienttypes.Height) *QueryChannelResponse {
	return &QueryChannelResponse{
		Channel:     &channel,
		Proof:       proof,
		ProofHeight: height,
	}
}

// NewQueryChannelClientStateResponse creates a newQueryChannelClientStateResponse instance
func NewQueryChannelClientStateResponse(identifiedClientState clienttypes.IdentifiedClientState, proof []byte, height clienttypes.Height) *QueryChannelClientStateResponse {
	return &QueryChannelClientStateResponse{
		IdentifiedClientState: &identifiedClientState,
		Proof:                 proof,
		ProofHeight:           height,
	}
}

// NewQueryChannelConsensusStateResponse creates a newQueryChannelConsensusStateResponse instance
func NewQueryChannelConsensusStateResponse(clientID string, anyConsensusState *codectypes.Any, consensusStateHeight exported.Height, proof []byte, height clienttypes.Height) *QueryChannelConsensusStateResponse {
	return &QueryChannelConsensusStateResponse{
		ConsensusState: anyConsensusState,
		ClientId:       clientID,
		Proof:          proof,
		ProofHeight:    height,
	}
}

// NewQueryPacketCommitmentResponse creates a new QueryPacketCommitmentResponse instance
func NewQueryPacketCommitmentResponse(
	commitment []byte, proof []byte, height clienttypes.Height,
) *QueryPacketCommitmentResponse {
	return &QueryPacketCommitmentResponse{
		Commitment:  commitment,
		Proof:       proof,
		ProofHeight: height,
	}
}

// NewQueryPacketAcknowledgementResponse creates a new QueryPacketAcknowledgementResponse instance
func NewQueryPacketAcknowledgementResponse(
	acknowledgement []byte, proof []byte, height clienttypes.Height,
) *QueryPacketAcknowledgementResponse {
	return &QueryPacketAcknowledgementResponse{
		Acknowledgement: acknowledgement,
		Proof:           proof,
		ProofHeight:     height,
	}
}

// NewQueryNextSequenceReceiveResponse creates a new QueryNextSequenceReceiveResponse instance
func NewQueryNextSequenceReceiveResponse(
	sequence uint64, proof []byte, height clienttypes.Height,
) *QueryNextSequenceReceiveResponse {
	return &QueryNextSequenceReceiveResponse{
		NextSequenceReceive: sequence,
		Proof:               proof,
		ProofHeight:         height,
	}
}
