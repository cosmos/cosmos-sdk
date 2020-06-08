package types

import (
	"strings"

	ics23 "github.com/confio/ics23/go"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
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

// StateResponse defines the client response for a client state query.
// It includes the commitment proof and the height of the proof.
type StateResponse struct {
	ClientState exported.ClientState        `json:"client_state" yaml:"client_state"`
	Proof       commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewClientStateResponse creates a new StateResponse instance.
func NewClientStateResponse(
	clientID string, clientState exported.ClientState, proofs []*ics23.CommitmentProof, height int64,
) StateResponse {
	return StateResponse{
		ClientState: clientState,
		Proof:       commitmenttypes.NewMerkleProof(proofs...),
		ProofPath:   commitmenttypes.NewMerklePath(append([]string{clientID}, strings.Split(host.ClientStatePath(), "/")...)),
		ProofHeight: uint64(height),
	}
}

// ConsensusStateResponse defines the client response for a Consensus state query.
// It includes the commitment proof and the height of the proof.
type ConsensusStateResponse struct {
	ConsensusState exported.ConsensusState     `json:"consensus_state" yaml:"consensus_state"`
	Proof          commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath      commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight    uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewConsensusStateResponse creates a new ConsensusStateResponse instance.
func NewConsensusStateResponse(
	clientID string, cs exported.ConsensusState, proofs []*ics23.CommitmentProof, height int64,
) ConsensusStateResponse {
	return ConsensusStateResponse{
		ConsensusState: cs,
		Proof:          commitmenttypes.NewMerkleProof(proofs...),
		ProofPath:      commitmenttypes.NewMerklePath(append([]string{clientID}, strings.Split(host.ClientStatePath(), "/")...)),
		ProofHeight:    uint64(height),
	}
}
