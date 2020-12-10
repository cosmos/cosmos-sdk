package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// NewQueryClientStateResponse creates a new QueryClientStateResponse instance.
func NewQueryClientStateResponse(
	clientStateAny *codectypes.Any, proof []byte, height Height,
) *QueryClientStateResponse {
	return &QueryClientStateResponse{
		ClientState: clientStateAny,
		Proof:       proof,
		ProofHeight: height,
	}
}

// NewQueryConsensusStateResponse creates a new QueryConsensusStateResponse instance.
func NewQueryConsensusStateResponse(
	consensusStateAny *codectypes.Any, proof []byte, height Height,
) *QueryConsensusStateResponse {
	return &QueryConsensusStateResponse{
		ConsensusState: consensusStateAny,
		Proof:          proof,
		ProofHeight:    height,
	}
}
