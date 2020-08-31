package types

import (
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewQueryClientStateResponse creates a new QueryClientStateResponse instance.
func NewQueryClientStateResponse(
	clientID string, clientStateAny *codectypes.Any, proof []byte, height Height,
) *QueryClientStateResponse {
	path := commitmenttypes.NewMerklePath(append([]string{clientID}, strings.Split(host.ClientStatePath(), "/")...))
	return &QueryClientStateResponse{
		ClientState: clientStateAny,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: height.EpochHeight,
	}
}

// NewQueryConsensusStateResponse creates a new QueryConsensusStateResponse instance.
func NewQueryConsensusStateResponse(
	clientID string, consensusStateAny *codectypes.Any, proof []byte, height Height,
) *QueryConsensusStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(clientID, host.ConsensusStatePath(height)), "/"))
	return &QueryConsensusStateResponse{
		ConsensusState: consensusStateAny,
		Proof:          proof,
		ProofPath:      path.Pretty(),
		ProofHeight:    height.EpochHeight,
	}
}
