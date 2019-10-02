package types

import (
	"bytes"
	"errors"

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
	Roots state.Indexer
	// Consensus state bytes
	ConsensusState state.Value
	// Boolean that states if the client is frozen when a misbehaviour proof is
	// submitted in the event of an equivocation.
	Frozen state.Boolean
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

// Update updates the consensus state and the state root from a provided header
func (state State) Update(ctx sdk.Context, header exported.Header) error {
	if !state.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return errors.New("client is frozen due to misbehaviour")
	}

	stored := state.GetConsensusState(ctx)
	updated, err := stored.CheckValidityAndUpdateState(header)
	if err != nil {
		return err
	}

	state.ConsensusState.Set(ctx, updated)
	state.Roots.Set(ctx, updated.GetHeight(), updated.GetRoot())

	return nil
}

// Freeze updates the state of the client in the event of a misbehaviour
func (state State) Freeze(ctx sdk.Context) error {
	if !state.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return errors.New("cannot freeze an already frozen client")
	}

	state.Frozen.Set(ctx, true)
	return nil
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

// exists verifies if the client exists or not
func (state State) exists(ctx sdk.Context) bool {
	return state.ConsensusState.Exists(ctx)
}

func (state State) prefix() []byte {
	return bytes.Split(state.ConsensusState.KeyBytes(), LocalRoot())[0]
}

type CounterState struct {
	id             string
	ConsensusState ics23.Value
}

func (counterState CounterState) Is(ctx sdk.Context, client exported.ConsensusState) bool {
	return counterState.ConsensusState.Is(ctx, client)
}
