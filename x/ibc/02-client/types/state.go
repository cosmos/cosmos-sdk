package types

import (
	"bytes"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// State is a type that represents the state of a client.
// Any actor holding the Stage can access on and modify that client information.
type State struct {
	// Client ID
	id string
	// Past state roots required to avoid race conditions between client updates
	// and proof-carrying transactions as defined in
	// https://github.com/cosmos/ics/tree/master/spec/ics-002-client-semantics#utilising-past-roots
	Roots state.Indexer `json:"roots" yaml:"roots"`
	// Consensus state bytes
	ConsensusState state.Value `json:"consensus_state" yaml:"consensus_state"`
	// Boolean that states if the client is frozen when a misbehaviour proof is
	// submitted in the event of an equivocation.
	Frozen bool `json:"frozen" yaml:"frozen"`
}

// NewState creates a new State instance
func NewState(id string, roots state.Indexer, consensusState state.Value) State {
	return State{
		id:             id,
		Roots:          roots,
		ConsensusState: consensusState,
		Frozen:         false,
	}
}

// ID returns the client identifier
func (state State) ID() string {
	return state.id
}

// GetConsensusState returns the consensus state
func (state State) GetConsensusState(ctx sdk.Context) (cs exported.ConsensusState) {
	state.ConsensusState.Get(ctx, &cs)
	return
}

// GetRoot returns the commitment root of the client at a given height
func (state State) GetRoot(ctx sdk.Context, height uint64) (root ics23.Root, err error) {
	err = state.Roots.GetSafe(ctx, height, &root)
	return
}

// Exists verifies if the client exists or not
func (state State) Exists(ctx sdk.Context) bool {
	return state.ConsensusState.Exists(ctx)
}

func (state State) RootCLI(q state.ABCIQuerier, height uint64) (res ics23.Root, proof merkle.Proof, err error) {
	root := state.Roots.Value(height)
	tmProof, err := root.Query(q, &res)
	proof = merkle.NewProofFromValue(tmProof, state.prefix(), root)
	return
}

func (state State) ConsensusStateCLI(q state.ABCIQuerier) (res exported.ConsensusState, proof merkle.Proof, err error) {
	tmproof, err := state.ConsensusState.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.ConsensusState)
	return
}

func (state State) FrozenCLI(q state.ABCIQuerier) (res bool, proof merkle.Proof, err error) {
	res, tmproof, err := state.Frozen.Query(q)
	proof = merkle.NewProofFromValue(tmproof, state.prefix(), state.Frozen)
	return
}

func (state State) prefix() []byte {
	return bytes.Split(state.ConsensusState.KeyBytes(), []byte(SubModuleName+"/"))[0]
}
