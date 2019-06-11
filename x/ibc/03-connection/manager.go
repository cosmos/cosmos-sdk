package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// XXX: rename remote to something else
// XXX: all code using external KVStore should be defer-recovered in case of missing proofs

type Manager struct {
	protocol state.Mapping

	client client.Manager

	self state.Indexer

	counterparty CounterpartyManager
}

func NewManager(protocol, free state.Base, clientman client.Manager) Manager {
	return Manager{
		protocol: state.NewMapping(protocol, []byte("/connection/")),

		client: clientman,

		self: state.NewIndexer(free, []byte("/connection/self/"), state.Dec),

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

// CONTRACT: client and remote must be filled by the caller
func (man Manager) object(id string) Object {
	return Object{
		id:          id,
		connection:  man.protocol.Value([]byte(id)),
		state:       state.NewEnum(man.protocol.Value([]byte(id + "/state"))),
		nexttimeout: state.NewInteger(man.protocol.Value([]byte(id+"/timeout")), state.Dec),

		self: man.self,
	}
}

// CONTRACT: client must be filled by the caller
func (man CounterpartyManager) object(id string) CounterObject {
	return CounterObject{
		id:          id,
		connection:  man.protocol.Value([]byte(id)),
		state:       commitment.NewEnum(man.protocol.Value([]byte(id + "/state"))),
		nexttimeout: commitment.NewInteger(man.protocol.Value([]byte(id+"/timeout")), state.Dec),
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
	obj.counterparty = man.counterparty.object(connection.Counterparty)
	obj.counterparty.client = man.counterparty.client.Query(connection.CounterpartyClient)
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
	obj.counterparty = man.counterparty.object(conn.Counterparty)
	obj.counterparty.client = man.counterparty.client.Query(conn.CounterpartyClient)
	return
}

// XXX: add HasProof() method to commitment.Store, and check it here
func (man CounterpartyManager) Query(id string) CounterObject {
	return man.object(id)
}

type NihiloObject Object

type Object struct {
	id          string
	connection  state.Value
	state       state.Enum
	nexttimeout state.Integer

	client client.Object

	counterparty CounterObject
	self         state.Indexer
}

type CounterObject struct {
	id          string
	connection  commitment.Value
	state       commitment.Enum
	nexttimeout commitment.Integer

	client client.CounterObject
}

func (obj Object) ID() string {
	return obj.id
}

func (obj Object) ClientID() string {
	return obj.client.ID()
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
}

// If there is no proof provided, assume not exists
func (obj CounterObject) exists(ctx sdk.Context) bool {
	/*
		// XXX
		if !obj.connection.Proven(ctx) {
			return false
		}

		return obj.connection.Exists(ctx)
	*/

	return false
}

func (obj Object) remove(ctx sdk.Context) {
	obj.connection.Delete(ctx)
	obj.state.Delete(ctx)
	obj.nexttimeout.Delete(ctx)
}

func (obj Object) assertSymmetric(ctx sdk.Context) error {
	if !obj.counterparty.connection.Is(ctx, obj.Value(ctx).Symmetry(obj.id)) {
		return errors.New("unexpected counterparty connection value")
	}

	return nil
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if uint64(ctx.BlockHeight()) > timeoutHeight {
		return errors.New("timeout")
	}

	return nil
}

func (obj Object) Value(ctx sdk.Context) (res Connection) {
	obj.connection.Get(ctx, &res)
	return
}

func (nobj NihiloObject) OpenInit(ctx sdk.Context, nextTimeoutHeight uint64) error {

	obj := Object(nobj)
	if obj.exists(ctx) {
		return errors.New("init on existing connection")
	}

	if !obj.state.Transit(ctx, Idle, Init) {
		return errors.New("init on non-idle connection")
	}

	// CONTRACT: OpenInit() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nexttimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

func (nobj NihiloObject) OpenTry(ctx sdk.Context, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {

	obj := Object(nobj)
	if !obj.state.Transit(ctx, Idle, OpenTry) {
		return errors.New("invalid state")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.state.Is(ctx, Init) {
		return errors.New("counterparty state not init")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil {
		return err
	}

	if !obj.counterparty.nexttimeout.Is(ctx, uint64(timeoutHeight)) {
		return errors.New("unexpected counterparty timeout value")
	}

	var expected client.ConsensusState
	obj.self.Get(ctx, expheight, &expected)
	if !obj.counterparty.client.Is(ctx, expected) {
		return errors.New("unexpected counterparty client value")
	}

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nexttimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

func (obj Object) OpenAck(ctx sdk.Context, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {

	if !obj.state.Transit(ctx, Init, Open) {
		return errors.New("ack on non-init connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.state.Is(ctx, OpenTry) {
		return errors.New("counterparty state not try")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil {
		return err
	}

	if !obj.counterparty.nexttimeout.Is(ctx, uint64(timeoutHeight)) {
		return errors.New("unexpected counterparty timeout value")
	}

	var expected client.ConsensusState
	obj.self.Get(ctx, expheight, &expected)
	if !obj.counterparty.client.Is(ctx, expected) {
		return errors.New("unexpected counterparty client value")
	}

	obj.nexttimeout.Set(ctx, uint64(nextTimeoutHeight))

	return nil
}

func (obj Object) OpenConfirm(ctx sdk.Context, timeoutHeight uint64) error {
	if !obj.state.Transit(ctx, OpenTry, Open) {
		return errors.New("confirm on non-try connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.state.Is(ctx, Open) {
		return errors.New("counterparty state not open")
	}

	err = obj.assertSymmetric(ctx)
	if err != nil {
		return err
	}

	if !obj.counterparty.nexttimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nexttimeout.Set(ctx, 0)

	return nil
}

func (obj Object) OpenTimeout(ctx sdk.Context) error {
	if !(uint64(obj.client.Value(ctx).GetHeight()) > obj.nexttimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	// XXX: double check if a user can bypass the verification logic somehow
	switch obj.state.Get(ctx) {
	case Init:
		if !obj.counterparty.connection.Is(ctx, nil) {
			return errors.New("counterparty connection exists")
		}
	case OpenTry:
		if !(obj.counterparty.state.Is(ctx, Init) ||
			obj.counterparty.exists(ctx)) {
			return errors.New("counterparty connection state not init")
		}
		// XXX: check if we need to verify symmetricity for timeout (already proven in OpenTry)
	case Open:
		if obj.counterparty.state.Is(ctx, OpenTry) {
			return errors.New("counterparty connection state not tryopen")
		}
	}

	obj.remove(ctx)

	return nil
}

func (obj Object) CloseInit(ctx sdk.Context, nextTimeout uint64) error {
	if !obj.state.Transit(ctx, Open, CloseTry) {
		return errors.New("closeinit on non-open connection")
	}

	obj.nexttimeout.Set(ctx, nextTimeout)

	return nil
}

func (obj Object) CloseTry(ctx sdk.Context, timeoutHeight, nextTimeoutHeight uint64) error {
	if !obj.state.Transit(ctx, Open, Closed) {
		return errors.New("closetry on non-open connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.state.Is(ctx, CloseTry) {
		return errors.New("unexpected counterparty state value")
	}

	if !obj.counterparty.nexttimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nexttimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

func (obj Object) CloseAck(ctx sdk.Context, timeoutHeight uint64) error {
	if !obj.state.Transit(ctx, CloseTry, Closed) {
		return errors.New("closeack on non-closetry connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.state.Is(ctx, Closed) {
		return errors.New("unexpected counterparty state value")
	}

	if !obj.counterparty.nexttimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nexttimeout.Set(ctx, 0)

	return nil
}

func (obj Object) CloseTimeout(ctx sdk.Context) error {
	if !(uint64(obj.client.Value(ctx).GetHeight()) > obj.nexttimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	// XXX: double check if the user can bypass the verification logic somehow
	switch obj.state.Get(ctx) {
	case CloseTry:
		if !obj.counterparty.state.Is(ctx, Open) {
			return errors.New("counterparty connection state not open")
		}
	case Closed:
		if !obj.counterparty.state.Is(ctx, CloseTry) {
			return errors.New("counterparty connection state not closetry")
		}
	}

	obj.state.Set(ctx, Open)
	obj.nexttimeout.Set(ctx, 0)

	return nil

}
