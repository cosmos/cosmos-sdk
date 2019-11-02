package types

import (
	"strings"

	tmtypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/tendermint/tendermint/crypto/merkle"
)

// query routes supported by the IBC client Querier
const (
	QueryClientState    = "client_state"
	QueryConsensusState = "consensus_state"
	QueryVerifiedRoot   = "roots"
)

// QueryClientStateParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/client_state'
// - 'custom/ibc/clients/<clientID>/consensus_state'
type QueryClientStateParams struct {
	ClientID string
}

// NewQueryClientStateParams creates a new QueryClientStateParams instance
func NewQueryClientStateParams(id string) QueryClientStateParams {
	return QueryClientStateParams{
		ClientID: id,
	}
}

// QueryCommitmentRootParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/roots/<height>'
type QueryCommitmentRootParams struct {
	ClientID string
	Height   uint64
}

// NewQueryCommitmentRootParams creates a new QueryCommitmentRootParams instance
func NewQueryCommitmentRootParams(id string, height uint64) QueryCommitmentRootParams {
	return QueryCommitmentRootParams{
		ClientID: id,
		Height:   height,
	}
}

type ConsensusStateResponse struct {
	ConsensusState tmtypes.ConsensusState
	Proof          commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath      commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight    uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

func NewConsensusStateResponse(
	clientID string, cs tmtypes.ConsensusState, proof *merkle.Proof, height int64,
) ConsensusStateResponse {
	return ConsensusStateResponse{
		ConsensusState: cs,
		Proof:          commitment.Proof{Proof: proof},
		ProofPath:      commitment.NewPath(strings.Split(ConsensusStatePath(clientID), "/")),
		ProofHeight:    uint64(height),
	}
}
