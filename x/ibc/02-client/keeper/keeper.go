package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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
func NewKeeper(mapping state.Mapping, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		mapping:   mapping.Prefix([]byte(types.SubModuleName + "/")),                          // "client/"
		codespace: sdk.CodespaceType(fmt.Sprintf("%s/%s", codespace, types.DefaultCodespace)), // "ibc/client"
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
		return types.State{}, sdkerrors.Wrap(err, "cannot create client")
	}

	// set the most recent state root and consensus state
	state.Roots.Set(ctx, cs.GetHeight(), cs.GetRoot())
	state.ConsensusState.Set(ctx, cs)
	return state, nil
}

// State returns a new client state with a given id
func (k Keeper) ClientState(id string) types.State {
	return types.NewState(
		id, // client ID
		k.mapping.Prefix([]byte(id+"/roots/")).Indexer(state.Dec), // commitment roots
		k.mapping.Value([]byte(id)),                               // consensus state
		k.mapping.Value([]byte(id+"/freeze")).Boolean(),           // client frozen
	)
}

// Query returns a client state that matches a given ID
func (k Keeper) Query(ctx sdk.Context, id string) (types.State, error) {
	state := k.ClientState(id)
	if !state.Exists(ctx) {
		return types.State{}, types.ErrClientExists(k.codespace)
	}
	return state, nil
}

// CheckMisbehaviourAndUpdateState checks for client misbehaviour and freezes the
// client if so.
func (k Keeper) CheckMisbehaviourAndUpdateState(ctx sdk.Context, state types.State, evidence exported.Evidence) error {
	var err error
	switch evidence.H1().Kind() {
	case exported.Tendermint:
		var tmEvidence tendermint.Evidence
		_, ok := evidence.(tendermint.Evidence)
		if !ok {
			return sdkerrors.Wrap(types.ErrInvalidConsensus(k.codespace), "consensus is not Tendermint")
		}
		err = tendermint.CheckMisbehaviour(tmEvidence)
	default:
		panic("unregistered consensus type")
	}

	if err != nil {
		return err
	}

	return k.Freeze(ctx, state)
}

// Update updates the consensus state and the state root from a provided header
func (k Keeper) Update(ctx sdk.Context, state types.State, header exported.Header) error {
	if !state.Exists(ctx) {
		panic("should not update nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return sdkerrors.Wrap(types.ErrClientFrozen(k.codespace), "cannot update client")
	}

	consensusState := state.GetConsensusState(ctx)
	consensusState, err := consensusState.CheckValidityAndUpdateState(header)
	if err != nil {
		return sdkerrors.Wrap(err, "cannot update client")
	}

	state.ConsensusState.Set(ctx, consensusState)
	state.Roots.Set(ctx, consensusState.GetHeight(), consensusState.GetRoot())
	return nil
}

// Freeze updates the state of the client in the event of a misbehaviour
func (k Keeper) Freeze(ctx sdk.Context, state types.State) error {
	if !state.Exists(ctx) {
		panic("should not freeze nonexisting client")
	}

	if state.Frozen.Get(ctx) {
		return sdkerrors.Wrap(types.ErrClientFrozen(k.codespace), "already frozen")
	}

	state.Frozen.Set(ctx, true)
	return nil
}
