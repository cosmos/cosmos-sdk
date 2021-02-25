package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	_ codectypes.UnpackInterfacesMessage = QueryChannelClientStateResponse{}
	_ codectypes.UnpackInterfacesMessage = QueryChannelConsensusStateResponse{}
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

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (qccsr QueryChannelClientStateResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return qccsr.IdentifiedClientState.UnpackInterfaces(unpacker)
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

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (qccsr QueryChannelConsensusStateResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(qccsr.ConsensusState, new(exported.ConsensusState))
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

// NewQueryPacketReceiptResponse creates a new QueryPacketReceiptResponse instance
func NewQueryPacketReceiptResponse(
	recvd bool, proof []byte, height clienttypes.Height,
) *QueryPacketReceiptResponse {
	return &QueryPacketReceiptResponse{
		Received:    recvd,
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
