package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	client "github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

type Manager struct {
	protocol     state.Mapping
	client       client.Manager
	counterparty CounterpartyManager
	path         merkle.Prefix
}

func NewManager(protocol state.Mapping, client client.Manager) Manager {
	return Manager{
		protocol:     protocol.Prefix(LocalRoot()),
		client:       client,
		counterparty: NewCounterpartyManager(protocol.Cdc()),
		path:         merkle.NewPrefix([][]byte{[]byte(protocol.StoreName())}, protocol.PrefixBytes()),
	}
}

type CounterpartyManager struct {
	protocol commitment.Mapping

	client client.CounterpartyManager
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	protocol := commitment.NewMapping(cdc, nil)

	return CounterpartyManager{
		protocol: protocol.Prefix(LocalRoot()),

		client: client.NewCounterpartyManager(cdc),
	}
}

type State struct {
	id string

	protocol   state.Mapping
	Connection state.Value
	Available  state.Boolean

	Kind state.String

	Client client.State

	path merkle.Prefix
}

// CONTRACT: client must be filled by the caller
func (man Manager) State(id string) State {
	return State{
		id:         id,
		protocol:   man.protocol.Prefix([]byte(id + "/")),
		Connection: man.protocol.Value([]byte(id)),
		Available:  man.protocol.Value([]byte(id + "/available")).Boolean(),
		Kind:       man.protocol.Value([]byte(id + "/kind")).String(),
		path:       man.path,
	}
}

type CounterState struct {
	id         string
	protocol   commitment.Mapping
	Connection commitment.Value
	Available  commitment.Boolean
	Kind       commitment.String
	Client     client.CounterState // nolint: unused
}

// CreateState creates a new CounterState instance.
// CONTRACT: client should be filled by the caller
func (man CounterpartyManager) CreateState(id string) CounterState {
	return CounterState{
		id:         id,
		protocol:   man.protocol.Prefix([]byte(id + "/")),
		Connection: man.protocol.Value([]byte(id)),
		Available:  man.protocol.Value([]byte(id + "/available")).Boolean(),
		Kind:       man.protocol.Value([]byte(id + "/kind")).String(),
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

func (state State) GetConnection(ctx sdk.Context) (res Connection) {
	state.Connection.Get(ctx, &res)
	return
}

func (state State) Sendable(ctx sdk.Context) bool {
	return kinds[state.Kind.Get(ctx)].Sendable
}

func (state State) Receivable(ctx sdk.Context) bool {
	return kinds[state.Kind.Get(ctx)].Receivable
}

func (state State) remove(ctx sdk.Context) {
	state.Connection.Delete(ctx)
	state.Available.Delete(ctx)
	state.Kind.Delete(ctx)
}

func (state State) exists(ctx sdk.Context) bool {
	return state.Connection.Exists(ctx)
}

func (man Manager) Cdc() *codec.Codec {
	return man.protocol.Cdc()
}

func (man Manager) create(ctx sdk.Context, id string, connection Connection, kind string) (state State, err error) {
	state = man.State(id)
	if state.exists(ctx) {
		err = errors.New("Stage already exists")
		return
	}
	state.Client, err = man.client.Query(ctx, connection.Client)
	if err != nil {
		return
	}
	state.Connection.Set(ctx, connection)
	state.Kind.Set(ctx, kind)
	return

}

// query() is used internally by the connection creators
// checks connection kind, doesn't check avilability
func (man Manager) query(ctx sdk.Context, id string, kind string) (state State, err error) {
	state = man.State(id)
	if !state.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	state.Client, err = man.client.Query(ctx, state.GetConnection(ctx).Client)
	if err != nil {
		return
	}

	if state.Kind.Get(ctx) != kind {
		err = errors.New("kind mismatch")
		return
	}

	return
}

func (man Manager) Query(ctx sdk.Context, id string) (state State, err error) {
	state = man.State(id)
	if !state.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	if !state.Available.Get(ctx) {
		err = errors.New("Stage not available")
		return
	}

	state.Client, err = man.client.Query(ctx, state.GetConnection(ctx).Client)
	return
}
