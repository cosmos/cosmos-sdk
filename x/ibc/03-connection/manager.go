package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/merkle"
)

type Manager struct {
	protocol state.Mapping

	client client.Manager

	counterparty CounterpartyManager

	path merkle.Path
}

func NewManager(protocol state.Mapping, client client.Manager) Manager {
	return Manager{
		protocol: protocol.Prefix(LocalRoot()),
		client:       client,
		counterparty: NewCounterpartyManager(protocol.Cdc()),
		path:         merkle.NewPath([][]byte{[]byte(protocol.StoreName())}, protocol.PrefixBytes()),
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

type Object struct {
	id string

	protocol   state.Mapping
	Connection state.Value
	Available  state.Boolean

	Kind state.String

	Client client.Object

	path merkle.Path
}

func (man Manager) Object(id string) Object {
	return Object{
		id: id,

		protocol:   man.protocol.Prefix([]byte(id + "/")),
		Connection: man.protocol.Value([]byte(id)),
		Available:  man.protocol.Value([]byte(id + "/available")).Boolean(),

		Kind: man.protocol.Value([]byte(id + "/kind")).String(),

		// CONTRACT: client must be filled by the caller

		path: man.path,
	}
}

type CounterObject struct {
	id string

	protocol   commitment.Mapping
	Connection commitment.Value
	Available  commitment.Boolean

	Kind commitment.String

	Client client.CounterObject // nolint: unused
}

func (man CounterpartyManager) Object(id string) CounterObject {
	return CounterObject{
		id:         id,
		protocol:   man.protocol.Prefix([]byte(id + "/")),
		Connection: man.protocol.Value([]byte(id)),
		Available:  man.protocol.Value([]byte(id + "/available")).Boolean(),

		Kind: man.protocol.Value([]byte(id + "/kind")).String(),

		// CONTRACT: client should be filled by the caller
	}
}

func (obj Object) Context(ctx sdk.Context, optpath commitment.Path, proofs []commitment.Proof) (sdk.Context, error) {
	if optpath == nil {
		optpath = obj.GetConnection(ctx).Path
	}

	store, err := commitment.NewStore(
		// TODO: proof root should be able to be obtained from the past
		obj.Client.GetConsensusState(ctx).GetRoot(),
		optpath,
		proofs,
	)
	if err != nil {
		return ctx, err
	}

	return commitment.WithStore(ctx, store), nil
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) GetConnection(ctx sdk.Context) (res Connection) {
	obj.Connection.Get(ctx, &res)
	return
}

func (obj Object) Sendable(ctx sdk.Context) bool {
	return kinds[obj.Kind.Get(ctx)].Sendable
}

func (obj Object) Receivble(ctx sdk.Context) bool {
	return kinds[obj.Kind.Get(ctx)].Receivable
}

func (obj Object) remove(ctx sdk.Context) {
	obj.Connection.Delete(ctx)
	obj.Available.Delete(ctx)
	obj.Kind.Delete(ctx)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.Connection.Exists(ctx)
}

func (man Manager) Cdc() *codec.Codec {
	return man.protocol.Cdc()
}

func (man Manager) create(ctx sdk.Context, id string, connection Connection, kind string) (obj Object, err error) {
	obj = man.Object(id)
	if obj.exists(ctx) {
		err = errors.New("Object already exists")
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
func (man Manager) query(ctx sdk.Context, id string, kind string) (obj Object, err error) {
	obj = man.Object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
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

func (man Manager) Query(ctx sdk.Context, id string) (obj Object, err error) {
	obj = man.Object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
		return
	}
	if !obj.Available.Get(ctx) {
		err = errors.New("Object not available")
		return
	}
	obj.Client, err = man.client.Query(ctx, obj.GetConnection(ctx).Client)
	return
}
