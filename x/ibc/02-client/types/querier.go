package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	tmtypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// query routes supported by the IBC client Querier
const (
	QueryAllClients     = "client_states"
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

// QueryAllClientsParams defines the parameters necessary for querying for all
// light client states.
type QueryAllClientsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAllClientsParams creates a new QueryAllClientsParams instance.
func NewQueryAllClientsParams(page, limit int) QueryAllClientsParams {
	return QueryAllClientsParams{
		Page:  page,
		Limit: limit,
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

// QueryCommitterParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/committers/<height>'
type QueryCommitterParams struct {
	ClientID string
	Height   uint64
}

// NewQueryCommitmentRootParams creates a new QueryCommitmentRootParams instance
func NewQueryCommitterParams(id string, height uint64) QueryCommitterParams {
	return QueryCommitterParams{
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

// CommitterResponse defines the client response for a committer query
// It includes the commitment proof and the height of the proof
type CommitterResponse struct {
	Committer   tmtypes.Committer `json:"committer" yaml:"committer"`
	Proof       commitment.Proof  `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path   `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64            `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewCommitterResponse creates a new CommitterResponse instance.
func NewCommitterResponse(
	clientID string, height uint64, committer tmtypes.Committer, proof *merkle.Proof, proofHeight int64,
) CommitterResponse {
	return CommitterResponse{
		Committer:   committer,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(RootPath(clientID, height), "/")),
		ProofHeight: uint64(proofHeight),
	}
}
