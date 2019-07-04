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

	table map[string]struct{}

	client client.Manager

	counterparty CounterpartyManager
}

type CounterpartyManager struct {
	protocol commitment.Mapping

	client client.CounterpartyManager
}

// CONTRACT: should be called on initialization time only
func (man Manager) Register(name string) PermissionedManager {
	if _, ok := man.table[name]; ok {
		panic("SubManager registered for existing name")
	}
	man.table[name] = struct{}{}

	return PermissionedManager{man, name}
}

func (man Manager) mapping(id string) state.Mapping {
	return man.protocol.Prefix([]byte(id))
}

func (man Manager) object(id string) Object {
	return Object{
		id:         id,
		connection: man.protocol.Value([]byte(id)),
		permission: state.NewString(man.protocol.Value([]byte(id + "/permission"))),

		mapping: man.protocol.Prefix([]byte(id)),

		// CONTRACT: client should be filled by the caller

		counterparty: man.counterparty.object(id),
	}
}

func (man CounterpartyManager) object(id string) CounterObject {
	return CounterObject{
		ID:         id,
		Connection: man.protocol.Value([]byte(id)),
		Permission: commitment.NewString(man.protocol.Value([]byte(id + "/permission"))),

		// CONTRACT: client should be filled by the caller

		Mapping: man.protocol.Prefix([]byte(id)),
	}
}

func (man Manager) Query(ctx sdk.Context, id string) (obj Object, err error) {
	obj = man.object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
		return
	}
	obj.client, err = man.client.Query(ctx, obj.Connection(ctx).GetClient())
	return
}

type PermissionedManager struct {
	man  Manager
	name string
}

func (man PermissionedManager) Cdc() *codec.Codec {
	return man.man.protocol.Cdc()
}

func (man PermissionedManager) object(id string) Object {
	return man.man.object(id)
}

func (man PermissionedManager) Create(ctx sdk.Context, id string, connection Connection) (obj Object, err error) {
	obj = man.object(id)
	if obj.exists(ctx) {
		err = errors.New("Object already exists")
		return
	}
	obj.connection.Set(ctx, connection)
	obj.permission.Set(ctx, man.name)
	obj.client, err = man.man.client.Query(ctx, connection.GetClient())
	return
}

func (man PermissionedManager) Query(ctx sdk.Context, id string) (obj Object, err error) {
	obj = man.object(id)
	if !obj.exists(ctx) {
		err = errors.New("Object not exists")
		return
	}
	if obj.Permission(ctx) != man.name {
		err = errors.New("Object created with different permission")
		return
	}
	obj.client, err = man.man.client.Query(ctx, obj.Connection(ctx).GetClient())
	return
}

type Object struct {
	id         string
	connection state.Value // Connection
	permission state.String

	mapping state.Mapping

	client client.Object

	counterparty CounterObject
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) Connection(ctx sdk.Context) (res Connection) {
	obj.connection.Get(ctx, &res)
	return
}

// TODO: it should not be exposed
func (obj Object) Permission(ctx sdk.Context) string {
	return obj.permission.Get(ctx)
}

func (obj Object) Mapping() state.Mapping {
	return obj.mapping
}

func (obj Object) Counterparty() CounterObject {
	return obj.counterparty
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
}

func (obj Object) Delete(ctx sdk.Context) {
	obj.connection.Delete(ctx)
	obj.permission.Delete(ctx)
}

type CounterObject struct {
	ID         string
	Connection commitment.Value
	Permission commitment.String

	Mapping commitment.Mapping

	Client client.CounterObject
}

// Flow(DEPRECATED):
// 1. a module calls SubManager.{Create, Query}()
// 2. SubManager calls StateManager.RequestBase(id string)
// 3-1. If id is not reserved, StateManager marks id as reserved and returns base
// 3-2. If id is reserved, StateManager checks if the SubManager is one who reserved
// 4. if okay, Manager return the base, which only Manager has access
// 5. SubManager returns the result of method call
