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

func NewManager(protocol state.Base, client client.Manager) Manager {
	return Manager{
		protocol:     state.NewMapping(protocol, ([]byte("/connection/"))),
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
	protocol := commitment.NewBase(cdc)

	return CounterpartyManager{
		protocol: commitment.NewMapping(protocol, []byte("/connection/")),

		client: client.NewCounterpartyManager(protocol),
	}
}

type Object struct {
	id string

	protocol   state.Mapping
	connection state.Value
	available  state.Boolean

	kind state.String

	client client.Object

	path merkle.Path
}

func (man Manager) object(id string) Object {
	return Object{
		id: id,

		protocol:   man.protocol.Prefix([]byte(id + "/")),
		connection: man.protocol.Value([]byte(id)),
		available:  state.NewBoolean(man.protocol.Value([]byte(id + "/available"))),

		kind: state.NewString(man.protocol.Value([]byte(id + "/kind"))),

		// CONTRACT: client must be filled by the caller

		path: man.path,
	}
}

type CounterObject struct {
	id string

	protocol   commitment.Mapping
	connection commitment.Value
	available  commitment.Boolean

	kind commitment.String

	client client.CounterObject // nolint: unused
}

func (man CounterpartyManager) object(id string) CounterObject {
	return CounterObject{
		id:         id,
		protocol:   man.protocol.Prefix([]byte(id + "/")),
		connection: man.protocol.Value([]byte(id)),
		available:  commitment.NewBoolean(man.protocol.Value([]byte(id + "/available"))),

		kind: commitment.NewString(man.protocol.Value([]byte(id + "/kind"))),

		// CONTRACT: client should be filled by the caller
	}
}

func (obj Object) Context(ctx sdk.Context, optpath commitment.Path, proofs []commitment.Proof) (sdk.Context, error) {
	if optpath == nil {
		optpath = obj.Connection(ctx).Path
	}

	store, err := commitment.NewStore(
		// TODO: proof root should be able to be obtained from the past
		obj.client.ConsensusState(ctx).GetRoot(),
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

func (obj Object) Connection(ctx sdk.Context) (res Connection) {
	obj.connection.Get(ctx, &res)
	return
}

func (obj Object) Available(ctx sdk.Context) bool {
	return obj.available.Get(ctx)
}

func (obj Object) Client() client.Object {
	return obj.client
}

func (obj Object) Sendable(ctx sdk.Context) bool {
	return kinds[obj.kind.Get(ctx)].Sendable
}

func (obj Object) Receivble(ctx sdk.Context) bool {
	return kinds[obj.kind.Get(ctx)].Receivable
}

func (obj Object) remove(ctx sdk.Context) {
	obj.connection.Delete(ctx)
	obj.available.Delete(ctx)
	obj.kind.Delete(ctx)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
}

func (man Manager) Cdc() *codec.Codec {
	return man.protocol.Cdc()
}

func (man Manager) create(ctx sdk.Context, id string, connection Connection, kind string) (obj Object, err error) {
	obj = man.object(id)
	if obj.exists(ctx) {
		err = errors.New("Object already exists")
		return
	}
	obj.client, err = man.client.Query(ctx, connection.Client)
	if err != nil {
		return
	}
	obj.connection.Set(ctx, connection)
	obj.kind.Set(ctx, kind)
	return

}

// query() is used internally by the connection creators
// checks connection kind, doesn't check avilability
func (man Manager) query(ctx sdk.Context, id string, kind string) (obj Object, err error) {
	obj = man.object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
		return
	}
	obj.client, err = man.client.Query(ctx, obj.Connection(ctx).Client)
	if err != nil {
		return
	}
	if obj.kind.Get(ctx) != kind {
		err = errors.New("kind mismatch")
		return
	}
	return
}

func (man Manager) Query(ctx sdk.Context, id string) (obj Object, err error) {
	obj = man.object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
		return

	}
	if !obj.Available(ctx) {
		err = errors.New("Object not available")
		return
	}
	obj.client, err = man.client.Query(ctx, obj.Connection(ctx).Client)
	return
}
