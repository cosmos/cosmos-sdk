package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// query routes supported by the IBC client Querier
const (
	QueryAllClients     = "client_states"
	QueryClientState    = "client_state"
	QueryConsensusState = "consensus_state"
)

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

// QueryCommitterParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/committers/<height>'
type QueryCommitterParams struct {
	ClientID string
	Height   uint64
}

// NewQueryCommitterParams creates a new QueryCommitmentRootParams instance
func NewQueryCommitterParams(id string, height uint64) QueryCommitterParams {
	return QueryCommitterParams{
		ClientID: id,
		Height:   height,
	}
}

// StateResponse defines the client response for a client state query.
// It includes the commitment proof and the height of the proof.
type StateResponse struct {
	ClientState exported.ClientState `json:"client_state" yaml:"client_state"`
	Proof       commitment.Proof     `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path      `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64               `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewClientStateResponse creates a new StateResponse instance.
func NewClientStateResponse(
	clientID string, clientState exported.ClientState, proof *merkle.Proof, height int64,
) StateResponse {
	return StateResponse{
		ClientState: clientState,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(ClientStatePath(clientID), "/")),
		ProofHeight: uint64(height),
	}
}

// ConsensusStateResponse defines the client response for a Consensus state query.
// It includes the commitment proof and the height of the proof.
type ConsensusStateResponse struct {
	ConsensusState exported.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
	Proof          commitment.Proof        `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath      commitment.Path         `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight    uint64                  `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewConsensusStateResponse creates a new ConsensusStateResponse instance.
func NewConsensusStateResponse(
	clientID string, cs exported.ConsensusState, proof *merkle.Proof, height int64,
) ConsensusStateResponse {
	return ConsensusStateResponse{
		ConsensusState: cs,
		Proof:          commitment.Proof{Proof: proof},
		ProofPath:      commitment.NewPath(strings.Split(ConsensusStatePath(clientID, uint64(height)), "/")),
		ProofHeight:    uint64(height),
	}
}

// CommitterResponse defines the client response for a committer query
// It includes the commitment proof and the height of the proof
type CommitterResponse struct {
	Committer   exported.Committer `json:"committer" yaml:"committer"`
	Proof       commitment.Proof   `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path    `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64             `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewCommitterResponse creates a new CommitterResponse instance.
func NewCommitterResponse(
	clientID string, height uint64, committer exported.Committer, proof *merkle.Proof, proofHeight int64,
) CommitterResponse {
	return CommitterResponse{
		Committer:   committer,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(CommitterPath(clientID, height), "/")),
		ProofHeight: uint64(proofHeight),
	}
}
