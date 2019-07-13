package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type State = byte

const (
	Idle State = iota
	Init
	OpenTry
	Open
	CloseTry
	Closed
)

const HandshakeKind = "handshake"

type Handshaker struct {
	man Manager

	counterparty CounterpartyHandshaker
}

// TODO: ocapify Manager; an actor who holds Manager
// should not be able to construct creaters from it
// or add Seal() method to Manager?
func NewHandshaker(man Manager) Handshaker {
	return Handshaker{
		man: man,

		counterparty: CounterpartyHandshaker{man.counterparty},
	}
}

type CounterpartyHandshaker struct {
	man CounterpartyManager
}

type HandshakeObject struct {
	Object

	state              state.Enum
	counterpartyClient state.String
	nextTimeout        state.Integer

	counterparty CounterHandshakeObject
}

type CounterHandshakeObject struct {
	CounterObject

	state              commitment.Enum
	counterpartyClient commitment.String
	nextTimeout        commitment.Integer
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) object(parent Object) HandshakeObject {
	return HandshakeObject{
		Object: parent,

		state:              state.NewEnum(man.man.protocol.Value([]byte(parent.id + "/state"))),
		counterpartyClient: state.NewString(man.man.protocol.Value([]byte(parent.id + "/counterpartyClient"))),
		nextTimeout:        state.NewInteger(man.man.protocol.Value([]byte(parent.id+"/timeout")), state.Dec),

		// CONTRACT: counterparty must be filled by the caller
	}
}

func (man CounterpartyHandshaker) object(id string) CounterHandshakeObject {
	return CounterHandshakeObject{
		CounterObject: man.man.object(id),

		state:              commitment.NewEnum(man.man.protocol.Value([]byte(id + "/state"))),
		counterpartyClient: commitment.NewString(man.man.protocol.Value([]byte(id + "/counterpartyClient"))),
		nextTimeout:        commitment.NewInteger(man.man.protocol.Value([]byte(id+"/timeout")), state.Dec),
	}
}

func (man Handshaker) create(ctx sdk.Context, id string, connection Connection, counterpartyClient string) (obj HandshakeObject, err error) {
	cobj, err := man.man.create(ctx, id, connection, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	obj.counterpartyClient.Set(ctx, counterpartyClient)
	obj.counterparty = man.counterparty.object(connection.Counterparty)
	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, id string) (obj HandshakeObject, err error) {
	cobj, err := man.man.query(ctx, id, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	obj.counterparty = man.counterparty.object(obj.Connection(ctx).Counterparty)
	return
}

func (obj HandshakeObject) State(ctx sdk.Context) byte {
	return obj.state.Get(ctx)
}

func (obj HandshakeObject) CounterpartyClient(ctx sdk.Context) string {
	return obj.counterpartyClient.Get(ctx)
}

func (obj HandshakeObject) Timeout(ctx sdk.Context) uint64 {
	return obj.nextTimeout.Get(ctx)
}

func (obj HandshakeObject) NextTimeout(ctx sdk.Context) uint64 {
	return obj.nextTimeout.Get(ctx)
}

func (obj HandshakeObject) remove(ctx sdk.Context) {
	obj.Object.remove(ctx)
	obj.state.Delete(ctx)
	obj.counterpartyClient.Delete(ctx)
	obj.nextTimeout.Delete(ctx)
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if uint64(ctx.BlockHeight()) > timeoutHeight {
		return errors.New("timeout")
	}

	return nil
}

// Using proofs: none
func (man Handshaker) OpenInit(ctx sdk.Context,
	id string, connection Connection, counterpartyClient string, nextTimeoutHeight uint64,
) (HandshakeObject, error) {
	// man.Create() will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)
	obj, err := man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return HandshakeObject{}, err
	}

	obj.nextTimeout.Set(ctx, nextTimeoutHeight)
	obj.state.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterparty.{connection,state,nextTimeout,counterpartyClient, client}
func (man Handshaker) OpenTry(ctx sdk.Context,
	connectionp, statep, timeoutp, counterpartyClientp /*, clientp*/ commitment.Proof,
	id string, connection Connection, counterpartyClient string, timeoutHeight, nextTimeoutHeight uint64,
) (obj HandshakeObject, err error) {
	obj, err = man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, connection.Path, connectionp, statep, timeoutp, counterpartyClientp)
	if err != nil {
		return
	}

	err = assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return
	}

	if !obj.counterparty.state.Is(ctx, Init) {
		err = errors.New("counterparty state not init")
		return
	}

	if !obj.counterparty.connection.Is(ctx, Connection{
		Client:       counterpartyClient,
		Counterparty: id,
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterparty connection")
		return
	}

	if !obj.counterparty.counterpartyClient.Is(ctx, connection.Client) {
		err = errors.New("counterparty client not match")
		return
	}

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		err = errors.New("unexpected counterparty timeout value")
		return
	}

	// TODO: commented out, need to check whether the stored client is compatible
	// make a separate module that manages recent n block headers
	// ref #4647
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

	obj.state.Set(ctx, OpenTry)
	obj.nextTimeout.Set(ctx, nextTimeoutHeight)

	return
}

// Using proofs: counterparty.{connection, state, timeout, counterpartyClient, client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	connectionp, statep, timeoutp, counterpartyClientp /*, clientp*/ commitment.Proof,
	id string /*expheight uint64, */, timeoutHeight, nextTimeoutHeight uint64,
) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, nil, connectionp, statep, timeoutp, counterpartyClientp)
	if err != nil {
		return
	}

	if !obj.state.Transit(ctx, Init, Open) {
		err = errors.New("ack on non-init connection")
		return
	}

	err = assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return
	}

	if !obj.counterparty.connection.Is(ctx, Connection{
		Client:       obj.CounterpartyClient(ctx),
		Counterparty: obj.ID(),
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterparty")
		return
	}

	if !obj.counterparty.state.Is(ctx, OpenTry) {
		err = errors.New("counterparty state not opentry")
		return
	}

	if !obj.counterparty.counterpartyClient.Is(ctx, obj.Connection(ctx).Client) {
		err = errors.New("counterparty client not match")
		return
	}

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		err = errors.New("unexpected counterparty timeout value")
		return
	}

	// TODO: implement in v1
	/*
		var expected client.ConsensusState
		// obj.self.Get(ctx, expheight, &expected)
		if !obj.counterparty.client.Is(ctx, expected) {
			// return errors.New("unexpected counterparty client value")
		}
	*/
	obj.available.Set(ctx, true)
	obj.nextTimeout.Set(ctx, nextTimeoutHeight)

	return
}

// Using proofs: counterparty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context,
	statep, timeoutp commitment.Proof,
	id string, timeoutHeight uint64) (obj HandshakeObject, err error) {

	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, nil, statep, timeoutp)
	if err != nil {
		return
	}

	if !obj.state.Transit(ctx, OpenTry, Open) {
		err = errors.New("confirm on non-try connection")
		return
	}

	err = assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return
	}

	if !obj.counterparty.state.Is(ctx, Open) {
		err = errors.New("counterparty state not open")
		return
	}

	if !obj.counterparty.nextTimeout.Is(ctx, timeoutHeight) {
		err = errors.New("unexpected counterparty timeout value")
		return
	}

	obj.available.Set(ctx, true)
	obj.nextTimeout.Set(ctx, 0)

	return
}

func (obj HandshakeObject) OpenTimeout(ctx sdk.Context) error {
	if !(obj.client.ConsensusState(ctx).GetHeight() > obj.nextTimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	switch obj.state.Get(ctx) {
	case Init:
		if !obj.counterparty.connection.Is(ctx, nil) {
			return errors.New("counterparty connection exists")
		}
	case OpenTry:
		if !(obj.counterparty.state.Is(ctx, Init) ||
			obj.counterparty.connection.Is(ctx, nil)) {
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

func (obj HandshakeObject) CloseInit(ctx sdk.Context, nextTimeout uint64) error {
	if !obj.state.Transit(ctx, Open, CloseTry) {
		return errors.New("closeinit on non-open connection")
	}

	obj.nextTimeout.Set(ctx, nextTimeout)

	return nil
}

func (obj HandshakeObject) CloseTry(ctx sdk.Context, timeoutHeight, nextTimeoutHeight uint64) error {
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

func (obj HandshakeObject) CloseAck(ctx sdk.Context, timeoutHeight uint64) error {
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

func (obj HandshakeObject) CloseTimeout(ctx sdk.Context) error {
	if !(obj.client.ConsensusState(ctx).GetHeight() > obj.nextTimeout.Get(ctx)) {
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
