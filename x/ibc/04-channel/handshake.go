package channel

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

type Handshaker struct {
	man Manager

	counterparty CounterpartyHandshaker
}

func (man Handshaker) Kind() string {
	return "handshake"
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

	state       state.Enum
	nextTimeout state.Integer

	counterparty CounterHandshakeObject
}

type CounterHandshakeObject struct {
	CounterObject

	state       commitment.Enum
	nextTimeout commitment.Integer
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) object(parent Object) HandshakeObject {
	prefix := parent.connid + "/channels/" + parent.chanid

	return HandshakeObject{
		Object: parent,

		state:       state.NewEnum(man.man.protocol.Value([]byte(prefix + "/state"))),
		nextTimeout: state.NewInteger(man.man.protocol.Value([]byte(prefix+"/timeout")), state.Dec),

		// CONTRACT: counterparty must be filled by the caller
	}
}

func (man CounterpartyHandshaker) object(connid, chanid string) CounterHandshakeObject {
	prefix := connid + "/channels/" + chanid

	return CounterHandshakeObject{
		CounterObject: man.man.object(connid, chanid),

		state:       commitment.NewEnum(man.man.protocol.Value([]byte(prefix + "/state"))),
		nextTimeout: commitment.NewInteger(man.man.protocol.Value([]byte(prefix+"/timeout")), state.Dec),
	}
}

func (man Handshaker) create(ctx sdk.Context, connid, chanid string, channel Channel) (obj HandshakeObject, err error) {
	cobj, err := man.man.create(ctx, connid, chanid, channel)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	counterconnid := obj.connection.Connection(ctx).Counterparty
	obj.counterparty = man.counterparty.object(counterconnid, channel.Counterparty)
	obj.counterparty.connection = man.counterparty.man.connection.Object(counterconnid)

	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, connid, chanid string) (obj HandshakeObject, err error) {
	cobj, err := man.man.query(ctx, connid, chanid)
	if err != nil {
		return
	}
	obj = man.object(cobj)
	channel := obj.Channel(ctx)
	counterconnid := obj.connection.Connection(ctx).Counterparty
	obj.counterparty = man.counterparty.object(counterconnid, channel.Counterparty)
	obj.counterparty.connection = man.counterparty.man.connection.Object(counterconnid)
	return
}

func (obj HandshakeObject) State(ctx sdk.Context) byte {
	return obj.state.Get(ctx)
}

func (obj HandshakeObject) Timeout(ctx sdk.Context) uint64 {
	return obj.nextTimeout.Get(ctx)
}

func (obj HandshakeObject) NextTimeout(ctx sdk.Context) uint64 {
	return obj.nextTimeout.Get(ctx)
}

/*
func (obj HandshakeObject) remove(ctx sdk.Context) {
	obj.Object.remove(ctx)
	obj.state.Delete(ctx)
	obj.counterpartyClient.Delete(ctx)
	obj.nextTimeout.Delete(ctx)
}
*/

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if uint64(ctx.BlockHeight()) > timeoutHeight {
		return errors.New("timeout")
	}

	return nil
}

// Using proofs: none
func (man Handshaker) OpenInit(ctx sdk.Context,
	connid, chanid string, channel Channel, nextTimeoutHeight uint64,
) (HandshakeObject, error) {
	// man.Create() will ensure
	// assert(get("channels/{identifier}") === null) and
	// set("channels/{identifier}", connection)
	obj, err := man.create(ctx, connid, chanid, channel)
	if err != nil {
		return HandshakeObject{}, err
	}

	obj.nextTimeout.Set(ctx, nextTimeoutHeight)
	obj.state.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterparty.{handshake,state,nextTimeout,clientid,client}
func (man Handshaker) OpenTry(ctx sdk.Context,
	connid, chanid string, channel Channel, timeoutHeight, nextTimeoutHeight uint64,
) (obj HandshakeObject, err error) {
	obj, err = man.create(ctx, connid, chanid, channel)
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

	if !obj.counterparty.channel.Is(ctx, Channel{
		Port:             channel.CounterpartyPort,
		Counterparty:     chanid,
		CounterpartyPort: "", // TODO
	}) {
		err = errors.New("wrong counterparty connection")
		return
	}

	if !obj.counterparty.nextTimeout.Is(ctx, uint64(timeoutHeight)) {
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

// Using proofs: counterparty.{handshake,state,nextTimeout,clientid,client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	connid, chanid string, timeoutHeight, nextTimeoutHeight uint64,
) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, connid, chanid)
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

	channel := obj.Channel(ctx)
	if !obj.counterparty.channel.Is(ctx, Channel{
		Port:             channel.CounterpartyPort,
		Counterparty:     chanid,
		CounterpartyPort: "", // TODO
	}) {
		err = errors.New("wrong counterparty")
		return
	}

	if !obj.counterparty.state.Is(ctx, OpenTry) {
		err = errors.New("counterparty state not opentry")
		return
	}

	if !obj.counterparty.nextTimeout.Is(ctx, uint64(timeoutHeight)) {
		err = errors.New("unexpected counterparty timeout value")
		return
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
	obj.available.Set(ctx, true)

	return
}

// Using proofs: counterparty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context, connid, chanid string, timeoutHeight uint64) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, connid, chanid)
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

// TODO
/*
func (obj HandshakeObject) OpenTimeout(ctx sdk.Context) error {
	if !(uint64(obj.connection.Client().ConsensusState(ctx).GetHeight()) > obj.nextTimeout.Get(ctx)) {
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
*/
