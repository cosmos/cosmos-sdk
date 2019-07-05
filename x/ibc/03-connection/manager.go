package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Manager struct {
	protocol state.Mapping

	client client.Manager

	counterparty CounterpartyManager
}

func NewManager(protocol state.Base, client client.Manager) Manager {
	return Manager{
		protocol:     state.NewMapping(protocol, ([]byte("/connection/"))),
		client:       client,
		counterparty: NewCounterpartyManager(protocol.Cdc()),
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

	protocol  state.Mapping
	clientid  state.String
	available state.Boolean

	kind state.String

	client client.Object
}

func (man Manager) object(id string) Object {
	return Object{
		id: id,

		protocol:  man.protocol.Prefix([]byte(id + "/")),
		clientid:  state.NewString(man.protocol.Value([]byte(id + "/client"))),
		available: state.NewBoolean(man.protocol.Value([]byte(id + "/available"))),

		kind: state.NewString(man.protocol.Value([]byte(id + "/kind"))),

		// CONTRACT: client must be filled by the caller
	}
}

type CounterObject struct {
	id string

	protocol  commitment.Mapping
	clientid  commitment.String
	available commitment.Boolean

	kind commitment.String

	client client.CounterObject
}

func (man CounterpartyManager) object(id string) CounterObject {
	return CounterObject{
		id:        id,
		protocol:  man.protocol.Prefix([]byte(id + "/")),
		clientid:  commitment.NewString(man.protocol.Value([]byte(id + "/client"))),
		available: commitment.NewBoolean(man.protocol.Value([]byte(id + "/available"))),

		kind: commitment.NewString(man.protocol.Value([]byte(id + "/kind"))),

		// CONTRACT: client should be filled by the caller
	}
}
func (obj Object) ID() string {
	return obj.id
}

func (obj Object) ClientID(ctx sdk.Context) string {
	return obj.clientid.Get(ctx)
}

func (obj Object) Available(ctx sdk.Context) bool {
	return obj.available.Get(ctx)
}

func (obj Object) Client() client.Object {
	return obj.client
}

func (obj Object) remove(ctx sdk.Context) {
	obj.clientid.Delete(ctx)
	obj.available.Delete(ctx)
	obj.kind.Delete(ctx)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.clientid.Exists(ctx)
}

func (man Manager) Cdc() *codec.Codec {
	return man.protocol.Cdc()
}

func (man Manager) create(ctx sdk.Context, id, clientid string, kind string) (obj Object, err error) {
	obj = man.object(id)
	if obj.exists(ctx) {
		err = errors.New("Object already exists")
		return
	}
	obj.clientid.Set(ctx, clientid)
	obj.kind.Set(ctx, kind)
	return
}

// query() is used internally by the connection creators
// checing the connection kind
func (man Manager) query(ctx sdk.Context, id string, kind string) (obj Object, err error) {
	obj, err = man.Query(ctx, id)
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
		err = errors.New("object not available")
		return
	}
	obj.client, err = man.client.Query(ctx, obj.ClientID(ctx))
	return
}
