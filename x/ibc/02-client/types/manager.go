package types

import (
	"bytes"
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// Any actor holding the Manager can access on and modify any client information
type Manager struct {
	protocol state.Mapping
}

// NewManager creates a new Manager instance
func NewManager(base state.Mapping) Manager {
	return Manager{
		protocol: base.Prefix(LocalRoot()),
	}
}

type CounterpartyManager struct {
	protocol ics23.Mapping
}

// NewCounterpartyManager creates a new CounterpartyManager instance
func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	return CounterpartyManager{
		protocol: ics23.NewMapping(cdc, LocalRoot()),
	}
}

/*
func (m Manager) RegisterKind(kind Kind, pred ValidityPredicate) Manager {
	if _, ok := m.pred[kind]; ok {
		panic("Kind already registered")
	}
	m.pred[kind] = pred
	return m
}
*/

func (m Manager) State(id string) State {
	return State{
		id:             id,
		Roots:          m.protocol.Prefix([]byte(id + "/roots/")).Indexer(state.Dec),
		ConsensusState: m.protocol.Value([]byte(id)),
		Frozen:         m.protocol.Value([]byte(id + "/freeze")).Boolean(),
	}
}

func (m Manager) Create(ctx sdk.Context, id string, cs exported.ConsensusState) (State, error) {
	state := m.State(id)
	if state.exists(ctx) {
		return State{}, errors.New("cannot create client on an existing id")
	}
	state.Roots.Set(ctx, cs.GetHeight(), cs.GetRoot())
	state.ConsensusState.Set(ctx, cs)
	return state, nil
}

func (m Manager) Query(ctx sdk.Context, id string) (State, error) {
	res := m.State(id)
	if !res.exists(ctx) {
		return State{}, errors.New("client doesn't exist")
	}
	return res, nil
}

func (m CounterpartyManager) State(id string) CounterState {
	return CounterState{
		id:             id,
		ConsensusState: m.protocol.Value([]byte(id)),
	}
}

func (m CounterpartyManager) Query(id string) CounterState {
	return m.State(id)
}

// Any actor holding the Stage can access on and modify that client information
type State struct {
	id             string
	Roots          state.Indexer
	ConsensusState state.Value // ConsensusState
	Frozen         state.Boolean
}

type CounterState struct {
	id             string
	ConsensusState ics23.Value
}

func (state State) ID() string {
	return state.id
}

func (state State) GetConsensusState(ctx sdk.Context) (res exported.ConsensusState) {
	state.ConsensusState.Get(ctx, &res)
	return
}

func (state State) GetRoot(ctx sdk.Context, height uint64) (res ics23.Root, err error) {
	err = state.Roots.GetSafe(ctx, height, &res)
	return
}

func (state CounterState) Is(ctx sdk.Context, client exported.ConsensusState) bool {
	return state.ConsensusState.Is(ctx, client)
}

func (state State) exists(ctx sdk.Context) bool {
	return state.ConsensusState.Exists(ctx)
}

func (state State) Update(ctx sdk.Context, header exported.Header) error {
	if !state.exists(ctx) {
		panic("should not update nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return errors.New("client is Frozen")
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

func (state State) Freeze(ctx sdk.Context) error {
	if !state.exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return errors.New("client is already Frozen")
	}

	state.Frozen.Set(ctx, true)

	return nil
}

func (state State) Delete(ctx sdk.Context) error {
	if !state.exists(ctx) {
		panic("should not delete nonexisting client")
	}

	if !state.Frozen.Get(ctx) {
		return errors.New("client is not Frozen")
	}

	state.ConsensusState.Delete(ctx)
	state.Frozen.Delete(ctx)

	return nil
}

func (state State) prefix() []byte {
	return bytes.Split(state.ConsensusState.KeyBytes(), LocalRoot())[0]
}

func (state State) RootCLI(q state.ABCIQuerier, height uint64) (res ics23.Root, proof merkle.Proof, err error) {
	root := state.Roots.Value(height)
	tmproof, err := root.Query(q, &res)
	proof = merkle.NewProofFromValue(tmproof, state.prefix(), root)
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
