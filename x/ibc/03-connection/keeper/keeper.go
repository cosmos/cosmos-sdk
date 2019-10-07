package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

// Keeper defines the IBC connection keeper
type Keeper struct {
	mapping      state.Mapping
	clientKeeper types.ClientKeeper
	
	counterparty CounterpartyManager
	path         merkle.Prefix
}

// NewKeeper creates a new IBC connection Keeper instance
func NewKeeper(mapping state.Mapping, ck types.ClientKeeper) Keeper {
	return Keeper{
		mapping:      mapping.Prefix([]byte(types.SubModuleName + "/")),
		clientKeeper: ck,
		counterparty: NewCounterpartyManager(mapping.Cdc()),
		path:         merkle.NewPrefix([][]byte{[]byte(mapping.StoreName())}, mapping.PrefixBytes()),
	}
}

type CounterpartyManager struct {
	mapping commitment.Mapping

	client ics02.CounterpartyManager
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	mapping := commitment.NewMapping(cdc, nil)

	return CounterpartyManager{
		mapping: mapping.Prefix([]byte(types.SubModuleName + "/")),

		client: ics02.NewCounterpartyManager(cdc),
	}
}

type State struct {
	id string

	mapping    state.Mapping
	Connection state.Value
	Available  state.Boolean

	Kind state.String

	Client ics02.State

	path merkle.Prefix
}

// CONTRACT: client must be filled by the caller
func (k Keeper) State(id string) State {
	return State{
		id:         id,
		mapping:    k.mapping.Prefix([]byte(id + "/")),
		Connection: k.mapping.Value([]byte(id)),
		Available:  k.mapping.Value([]byte(id + "/available")).Boolean(),
		Kind:       k.mapping.Value([]byte(id + "/kind")).String(),
		path:       k.path,
	}
}

type CounterState struct {
	id         string
	mapping    commitment.Mapping
	Connection commitment.Value
	Available  commitment.Boolean
	Kind       commitment.String
	Client     ics02.CounterState // nolint: unused
}

// CreateState creates a new CounterState instance.
// CONTRACT: client should be filled by the caller
func (k CounterpartyManager) CreateState(id string) CounterState {
	return CounterState{
		id:         id,
		mapping:    k.mapping.Prefix([]byte(id + "/")),
		Connection: k.mapping.Value([]byte(id)),
		Available:  k.mapping.Value([]byte(id + "/available")).Boolean(),
		Kind:       k.mapping.Value([]byte(id + "/kind")).String(),
	}
}

func (state State) Context(ctx sdk.Context, height uint64, proofs []commitment.Proof) (sdk.Context, error) {
	root, err := state.Client.GetRoot(ctx, height)
	if err != nil {
		return ctx, err
	}

	store, err := commitment.NewStore(
		root,
		state.GetConnection(ctx).Path,
		proofs,
	)
	if err != nil {
		return ctx, err
	}

	return commitment.WithStore(ctx, store), nil
}

func (state State) ID() string {
	return state.id
}

func (state State) GetConnection(ctx sdk.Context) (res types.ConnectionEnd) {
	state.Connection.Get(ctx, &res)
	return
}

func (state State) Sendable(ctx sdk.Context) bool {
	return kinds[state.Kind.Get(ctx)].Sendable
}

func (state State) Receivable(ctx sdk.Context) bool {
	return kinds[state.Kind.Get(ctx)].Receivable
}

func (state State) exists(ctx sdk.Context) bool {
	return state.Connection.Exists(ctx)
}

func (k Keeper) Cdc() *codec.Codec {
	return k.mapping.Cdc()
}

func (k Keeper) create(ctx sdk.Context, id string, connection Connection, kind string) (state State, err error) {
	state = k.State(id)
	if state.exists(ctx) {
		err = errors.New("Stage already exists")
		return
	}
	state.Client, err = k.client.Query(ctx, connection.Client)
	if err != nil {
		return
	}
	state.Connection.Set(ctx, connection)
	state.Kind.Set(ctx, kind)
	return

}

// query() is used internally by the connection creators
// checks connection kind, doesn't check avilability
func (k Keeper) query(ctx sdk.Context, id string, kind string) (state State, err error) {
	state = k.State(id)
	if !state.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	state.Client, err = k.client.Query(ctx, state.GetConnection(ctx).Client)
	if err != nil {
		return
	}

	if state.Kind.Get(ctx) != kind {
		err = errors.New("kind mismatch")
		return
	}

	return
}

func (k Keeper) Query(ctx sdk.Context, id string) (state State, err error) {
	state = k.State(id)
	if !state.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	if !state.Available.Get(ctx) {
		err = errors.New("Stage not available")
		return
	}

	state.Client, err = k.client.Query(ctx, state.GetConnection(ctx).Client)
	return
}
