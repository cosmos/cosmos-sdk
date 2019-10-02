package keeper

import (
	"errors"
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// Keeper represents a type that grants read and write permissions to any client
// state information
type Keeper struct {
	mapping   state.Mapping
	codespace sdk.CodespaceType
}

// NewKeeper creates a new NewKeeper instance
func NewKeeper(mapping state.Mapping) Keeper {
	return Keeper{
		mapping: mapping.Prefix([]byte(types.SubModuleName + "/")),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s/%s", ibctypes.ModuleName, types.SubModuleName))
}

// CreateClient creates a new client state and populates it with a given consensus state
func (k Keeper) CreateClient(ctx sdk.Context, id string, cs exported.ConsensusState) (types.State, error) {
	state, err := k.Query(ctx, id)
	if err == nil {
		return types.State{}, errors.New("cannot create client on an existing id")
	}

	// set the most recent state root and consensus state
	state.Roots.Set(ctx, cs.GetHeight(), cs.GetRoot())
	state.ConsensusState.Set(ctx, cs)
	return state, nil
}

// State returnts a new client state with a given id
func (k Keeper) State(id string) types.State {
	return types.NewState(
		id, // client ID
		k.mapping.Prefix([]byte(id+"/roots/")).Indexer(state.Dec), // commitment roots
		k.mapping.Value([]byte(id)),                               // consensus state
		k.mapping.Value([]byte(id+"/freeze")).Boolean(),           // client frozen
	)
}

// Query returns a client state that matches a given ID
func (k Keeper) Query(ctx sdk.Context, id string) (types.State, error) {
	state := k.State(id)
	if !state.Exists(ctx) {
		return types.State{}, errors.New("client doesn't exist")
	}
	return state, nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, state types.State, evidence exported.Evidence) error {
	var err error
	switch evidence.H1().Kind() {
	case exported.Tendermint:
		err = tendermint.CheckMisbehaviour(evidence)
	default:
		panic("unregistered consensus type")
	}

	if err != nil {
		return err
	}

	return state.Freeze(ctx)
}
