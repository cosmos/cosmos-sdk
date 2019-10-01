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

func (obj State) Context(ctx sdk.Context, height uint64, proofs []commitment.Proof) (sdk.Context, error) {
	root, err := obj.Client.GetRoot(ctx, height)
	if err != nil {
		return ctx, err
	}

	store, err := commitment.NewStore(
		root,
		obj.GetConnection(ctx).Path,
		proofs,
	)
	if err != nil {
		return ctx, err
	}

	return commitment.WithStore(ctx, store), nil
}

func (obj State) ID() string {
	return obj.id
}

func (obj State) GetConnection(ctx sdk.Context) (res Connection) {
	obj.Connection.Get(ctx, &res)
	return
}

func (obj State) Sendable(ctx sdk.Context) bool {
	return kinds[obj.Kind.Get(ctx)].Sendable
}

func (obj State) Receivable(ctx sdk.Context) bool {
	return kinds[obj.Kind.Get(ctx)].Receivable
}

func (obj State) remove(ctx sdk.Context) {
	obj.Connection.Delete(ctx)
	obj.Available.Delete(ctx)
	obj.Kind.Delete(ctx)
}

func (obj State) exists(ctx sdk.Context) bool {
	return obj.Connection.Exists(ctx)
}

func (man Manager) Cdc() *codec.Codec {
	return man.protocol.Cdc()
}

func (man Manager) create(ctx sdk.Context, id string, connection Connection, kind string) (obj State, err error) {
	obj = man.State(id)
	if obj.exists(ctx) {
		err = errors.New("Stage already exists")
		return
	}
	obj.Client, err = man.client.Query(ctx, connection.Client)
	if err != nil {
		return
	}
	obj.Connection.Set(ctx, connection)
	obj.Kind.Set(ctx, kind)
	return

}

// query() is used internally by the connection creators
// checks connection kind, doesn't check avilability
func (man Manager) query(ctx sdk.Context, id string, kind string) (obj State, err error) {
	obj = man.State(id)
	if !obj.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	obj.Client, err = man.client.Query(ctx, obj.GetConnection(ctx).Client)
	if err != nil {
		return
	}

	if obj.Kind.Get(ctx) != kind {
		err = errors.New("kind mismatch")
		return
	}

	return
}

func (man Manager) Query(ctx sdk.Context, id string) (obj State, err error) {
	obj = man.State(id)
	if !obj.exists(ctx) {
		err = errors.New("Stage not exists")
		return
	}

	if !obj.Available.Get(ctx) {
		err = errors.New("Stage not available")
		return
	}

	obj.Client, err = man.client.Query(ctx, obj.GetConnection(ctx).Client)
	return
}
