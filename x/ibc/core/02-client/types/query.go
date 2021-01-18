package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

var (
	_ codectypes.UnpackInterfacesMessage = QueryClientStateResponse{}
	_ codectypes.UnpackInterfacesMessage = QueryClientStatesResponse{}
	_ codectypes.UnpackInterfacesMessage = QueryConsensusStateResponse{}
	_ codectypes.UnpackInterfacesMessage = QueryConsensusStatesResponse{}
)

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (qcsr QueryClientStatesResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	for _, cs := range qcsr.ClientStates {
		if err := cs.UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

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

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (qcsr QueryClientStateResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(qcsr.ClientState, new(exported.ClientState))
}

// unpackinterfaces implements unpackinterfacesmesssage.unpackinterfaces
func (qcsr queryconsensusstatesresponse) unpackinterfaces(unpacker codectypes.anyunpacker) error {
	for _, cs := range qcsr.consensusstates {
		if err := cs.unpackinterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
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

// UnpackInterfaces implements UnpackInterfacesMesssage.UnpackInterfaces
func (qcsr QueryConsensusStateResponse) UnpackInterfaces(unpacker codectypes.AnyUnpacker) error {
	return unpacker.UnpackAny(qcsr.ConsensusState, new(exported.ConsensusState))
}
