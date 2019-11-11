package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	tmtypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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

// StateResponse defines the client response for a client state query.
// It includes the commitment proof and the height of the proof.
type StateResponse struct {
	ClientState State            `json:"client_state" yaml:"client_state"`
	Proof       commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewClientStateResponse creates a new StateResponse instance.
func NewClientStateResponse(
	clientID string, clientState State, proof *merkle.Proof, height int64,
) StateResponse {
	return StateResponse{
		ClientState: clientState,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(ConsensusStatePath(clientID), "/")),
		ProofHeight: uint64(height),
	}
}

// ConsensusStateResponse defines the client response for a Consensus state query.
// It includes the commitment proof and the height of the proof.
type ConsensusStateResponse struct {
	ConsensusState tmtypes.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
	Proof          commitment.Proof       `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath      commitment.Path        `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight    uint64                 `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewConsensusStateResponse creates a new ConsensusStateResponse instance.
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

// RootResponse defines the client response for a commitment root query.
// It includes the commitment proof and the height of the proof.
type RootResponse struct {
	Root        commitment.Root  `json:"root" yaml:"root"`
	Proof       commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewRootResponse creates a new RootResponse instance.
func NewRootResponse(
	clientID string, height uint64, root commitment.Root, proof *merkle.Proof, proofHeight int64,
) RootResponse {
	return RootResponse{
		Root:        root,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(RootPath(clientID, height), "/")),
		ProofHeight: uint64(proofHeight),
	}
}
