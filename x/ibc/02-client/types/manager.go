package types

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// Manager represents a type that grants read and write permissions to any client
// state information
type Manager struct {
	protocol state.Mapping
}

// NewManager creates a new Manager instance
func NewManager(mapping state.Mapping) Manager {
	return Manager{
		protocol: mapping.Prefix(LocalRoot()),
	}
}

// CreateClient creates a new client state and populates it with a given consensus state
func (m Manager) CreateClient(ctx sdk.Context, id string, cs exported.ConsensusState) (State, error) {
	state := m.State(id)
	if state.exists(ctx) {
		return State{}, errors.New("cannot create client on an existing id")
	}

	// set the most recent state root and consensus state
	state.Roots.Set(ctx, cs.GetHeight(), cs.GetRoot())
	state.ConsensusState.Set(ctx, cs)
	return state, nil
}

// State returnts a new client state with a given id
func (m Manager) State(id string) State {
	return State{
		id:             id,
		Roots:          m.protocol.Prefix([]byte(id + "/roots/")).Indexer(state.Dec),
		ConsensusState: m.protocol.Value([]byte(id)),
		Frozen:         m.protocol.Value([]byte(id + "/freeze")).Boolean(),
	}
}

// Query returns a client state that matches a given ID
func (m Manager) Query(ctx sdk.Context, id string) (State, error) {
	res := m.State(id)
	if !res.exists(ctx) {
		return State{}, errors.New("client doesn't exist")
	}
	return res, nil
}
