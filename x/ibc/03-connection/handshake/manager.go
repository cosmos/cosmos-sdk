package handshake

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Manager struct {
	connection.PermissionedManager

	counterparty CounterpartyManager
}

func NewManager(man connection.PermissionedManager) Manager {
	return Manager{
		PermissionedManager: man,

		counterparty: NewCounterpartyManager(man.Cdc()),
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
func (man Manager) object(parent connection.Object) Object {
	return Object{
		Object: parent,

		nextTimeout: state.NewInteger(parent.Mapping().Value([]byte("/timeout")), state.Dec),

		counterparty: man.counterparty.object(parent.Counterparty()),
	}
}

func (man CounterpartyManager) object(parent connection.CounterObject) CounterObject {
	return CounterObject{
		CounterObject: parent,

		nextTimeout: commitment.NewInteger(parent.Mapping.Value([]byte("/timeout")), state.Dec),
	}
}

func (man Manager) Create(ctx sdk.Context, id string, connection Connection) (obj Object, err error) {
	cobj, err := man.PermissionedManager.Create(ctx, id, connection)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	return obj, nil
}

func (man Manager) query(ctx sdk.Context, id string) (obj Object, err error) {
	cobj, err := man.PermissionedManager.Query(ctx, id)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	return
}

type Object struct {
	connection.Object

	nextTimeout state.Integer

	client client.Object

	counterparty CounterObject
}

type CounterObject struct {
	connection.CounterObject

	nextTimeout commitment.Integer

	client client.CounterObject
}

func (obj Object) NextTimeout(ctx sdk.Context) uint64 {
	return obj.nextTimeout.Get(ctx)
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

// TODO: make it private
func (obj Object) Delete(ctx sdk.Context) {
	obj.Object.Remove(ctx)
	obj.nextTimeout.Delete(ctx)
}

func (obj Object) assertSymmetric(ctx sdk.Context) error {
	if !obj.counterparty.connection.Is(ctx, obj.Connection(ctx).Symmetry(obj.id)) {
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

// Using proofs: none
func (obj Object) OpenInit(ctx sdk.Context, nextTimeoutHeight uint64) error {
	if !obj.state.Transit(ctx, Idle, Init) {
		return errors.New("init on non-idle connection")
	}

	// CONTRACT: OpenInit() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nextTimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

// Using proofs: counterparty.{connection,state,nextTimeout,client}
func (obj Object) OpenTry(ctx sdk.Context, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {
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

	if !obj.counterparty.nextTimeout.Is(ctx, uint64(timeoutHeight)) {
		return errors.New("unexpected counterparty timeout value")
	}

	// TODO: commented out, need to check whether the stored client is compatible
	// make a separate module that manages recent n block headers
	/*
		var expected client.ConsensusState
		obj.self.Get(ctx, expheight, &expected)
		if !obj.counterparty.client.Is(ctx, expected) {
			return errors.New("unexpected counterparty client value")
		}
	*/

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.nextTimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

// Using proofs: counterparty.{connection,state, nextTimeout}
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

	if !obj.counterparty.nextTimeout.Is(ctx, uint64(timeoutHeight)) {
		return errors.New("unexpected counterparty timeout value")
	}

	// TODO: commented out, implement in v1
	/*
		var expected client.ConsensusState
		obj.self.Get(ctx, expheight, &expected)
		if !obj.counterparty.client.Is(ctx, expected) {
			return errors.New("unexpected counterparty client value")
		}
	*/

	obj.nextTimeout.Set(ctx, uint64(nextTimeoutHeight))

	return nil
}

// Using proofs: counterparty.{connection,state, nextTimeout}
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

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nextTimeout.Set(ctx, 0)

	return nil
}

func (obj Object) OpenTimeout(ctx sdk.Context) error {
	if !(uint64(obj.client.ConsensusState(ctx).GetHeight()) > obj.nextTimeout.Get(ctx)) {
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

	obj.nextTimeout.Set(ctx, nextTimeout)

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

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nextTimeout.Set(ctx, nextTimeoutHeight)

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

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.nextTimeout.Set(ctx, 0)

	return nil
}

func (obj Object) CloseTimeout(ctx sdk.Context) error {
	if !(uint64(obj.client.ConsensusState(ctx).GetHeight()) > obj.nextTimeout.Get(ctx)) {
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
	obj.nextTimeout.Set(ctx, 0)

	return nil

}
