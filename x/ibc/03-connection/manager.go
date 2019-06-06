package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// XXX: rename remote to something else
// XXX: all code using external KVStore should be defer-recovered in case of missing proofs

type Manager struct {
	protocol mapping.Mapping

	client client.Manager

	// CONTRACT: remote/self should not be used when remote
	remote *Manager
	self   mapping.Indexer
}

func NewManager(protocol, free mapping.Base, client client.Manager) Manager {
	return Manager{
		protocol: mapping.NewMapping(protocol, []byte("/")),

		client: client,

		self: mapping.NewIndexer(free, []byte("/self"), mapping.Dec),
	}
}

// TODO: return newtyped manager
func NewRemoteManager(protocol mapping.Base, client client.Manager) Manager {
	return NewManager(protocol, mapping.EmptyBase(), client)
}

func (man Manager) Exec(remote Manager, fn func(Manager)) {
	fn(Manager{
		protocol: man.protocol,
		client:   man.client,
		self:     man.self,
		remote:   &remote,
	})
}

// CONTRACT: client and remote must be filled by the caller
func (man Manager) object(id string) Object {
	return Object{
		id:          id,
		connection:  man.protocol.Value([]byte(id)),
		state:       man.protocol.Value([]byte(id + "/state")).Enum(),
		nexttimeout: man.protocol.Value([]byte(id + "/timeout")).Integer(),

		self: man.self,
	}
}

func (man Manager) Create(ctx sdk.Context, id string, connection Connection) (nobj NihiloObject, err error) {
	obj := man.object(id)
	if obj.exists(ctx) {
		err = errors.New("connection already exists for the provided id")
		return
	}
	obj.connection.Set(ctx, connection)
	obj.client, err = man.client.Query(ctx, connection.Client)
	if err != nil {
		return
	}
	remote := man.remote.object(connection.Counterparty)
	obj.remote = &remote
	return NihiloObject(obj), nil
}

func (man Manager) Query(ctx sdk.Context, key string) (obj Object, err error) {
	obj = man.object(key)
	if !obj.exists(ctx) {
		return Object{}, errors.New("connection not exists for the provided id")
	}
	conn := obj.Value(ctx)
	obj.client, err = man.client.Query(ctx, conn.Client)
	if err != nil {
		return
	}
	remote := man.remote.object(conn.Counterparty)
	obj.remote = &remote
	return
}

type NihiloObject Object

type Object struct {
	id          string
	connection  mapping.Value
	state       mapping.Enum
	nexttimeout mapping.Integer

	client client.Object

	// CONTRACT: remote/self should not be used when remote
	remote *Object
	self   mapping.Indexer
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
}

func (obj Object) remove(ctx sdk.Context) {
	obj.connection.Delete(ctx)
	obj.state.Delete(ctx)
	obj.nexttimeout.Delete(ctx)
}

func (obj Object) assertSymmetric(ctx sdk.Context) error {
	if !obj.Value(ctx).Symmetric(obj.id, obj.remote.Value(ctx)) {
		return errors.New("unexpected counterparty connection value")
	}

	return nil
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if ctx.BlockHeight() > int64(timeoutHeight) {
		return errors.New("timeout")
	}

	return nil
}

func (obj Object) Value(ctx sdk.Context) (res Connection) {
	obj.connection.Get(ctx, &res)
	return
}

func (nobj NihiloObject) OpenInit(ctx sdk.Context, nextTimeoutHeight uint64) error {
	keylog := commitment.IsKeyLog(ctx)

	obj := Object(nobj)
	if obj.exists(ctx) && !keylog {
		return errors.New("init on existing connection")
	}

	if !obj.state.Transit(ctx, Idle, Init) && !keylog {
		return errors.New("init on non-idle connection")
	}

	// CONTRACT: OpenInit() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nexttimeout.Set(ctx, int64(nextTimeoutHeight))

	return nil
}

func (nobj NihiloObject) OpenTry(ctx sdk.Context, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {
	keylog := commitment.IsKeyLog(ctx)

	obj := Object(nobj)
	if !obj.state.Transit(ctx, Idle, OpenTry) {
		return errors.New("invalid state")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if obj.remote.state.Get(ctx) != Init && !keylog {
		return errors.New("counterparty state not init")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil && !keylog {
		return err
	}

	if obj.remote.nexttimeout.Get(ctx) != int64(timeoutHeight) && !keylog {
		return errors.New("unexpected counterparty timeout value")
	}

	var expected client.Client
	obj.self.Get(ctx, expheight, &expected)
	if !client.Equal(obj.remote.client.Value(ctx), expected) && !keylog {
		return errors.New("unexpected counterparty client value")
	}

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nexttimeout.Set(ctx, int64(nextTimeoutHeight))

	return nil
}

func (obj Object) OpenAck(ctx sdk.Context, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {
	keylog := commitment.IsKeyLog(ctx)

	if !obj.state.Transit(ctx, Init, Open) {
		return errors.New("ack on non-init connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if obj.remote.state.Get(ctx) != OpenTry && !keylog {
		return errors.New("counterparty state not try")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil {
		return err
	}

	if obj.remote.nexttimeout.Get(ctx) != int64(timeoutHeight) && !keylog {
		return errors.New("unexpected counterparty timeout value")
	}

	var expected client.Client
	obj.self.Get(ctx, expheight, &expected)
	if !client.Equal(obj.remote.client.Value(ctx), expected) && !keylog {
		return errors.New("unexpected counterparty client value")
	}

	obj.nexttimeout.Set(ctx, int64(nextTimeoutHeight))

	return nil
}

func (obj Object) OpenConfirm(ctx sdk.Context, timeoutHeight uint64) error {
	keylog := commitment.IsKeyLog(ctx)

	if !obj.state.Transit(ctx, OpenTry, Open) {
		return errors.New("confirm on non-try connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if obj.remote.state.Get(ctx) != Open && !keylog {
		return errors.New("counterparty state not open")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil && !keylog {
		return err
	}

	if obj.remote.nexttimeout.Get(ctx) != int64(timeoutHeight) && !keylog {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nexttimeout.Set(ctx, 0)

	return nil
}

func (obj Object) OpenTimeout(ctx sdk.Context) error {
	keylog := commitment.IsKeyLog(ctx)

	if !(obj.client.Value(ctx).GetBase().GetHeight() > obj.nexttimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	switch obj.state.Get(ctx) {
	case Init:
		// XXX: check if exists compatible with remote KVStore
		if obj.remote.exists(ctx) && !keylog {
			return errors.New("counterparty connection existw")
		}
	case OpenTry:
		if !(obj.remote.state.Get(ctx) == Init ||
			obj.remote.exists(ctx)) && !keylog {
			return errors.New("counterparty connection state not init")
		}
		// XXX: check if we need to verify symmetricity for timeout (already proven in OpenTry)
	case Open:
		if obj.remote.state.Get(ctx) == OpenTry && !keylog {
			return errors.New("counterparty connection state not tryopen")
		}
	}

	obj.remove(ctx)

	return nil
}

func (obj Object) CloseInit(ctx sdk.Context) error {
	return nil // XXX
}
