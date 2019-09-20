package channel

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
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
	Manager

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
		Manager: man,

		counterparty: CounterpartyHandshaker{man.counterparty},
	}
}

type CounterpartyHandshaker struct {
	man CounterpartyManager
}

type HandshakeObject struct {
	Object

	State state.Enum

	counterparty CounterHandshakeObject
}

type CounterHandshakeObject struct {
	CounterObject

	State commitment.Enum
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) object(parent Object) HandshakeObject {
	prefix := parent.portid + "/channels/" + parent.chanid

	return HandshakeObject{
		Object: parent,

		State: man.protocol.Value([]byte(prefix + "/state")).Enum(),

		counterparty: man.counterparty.object(parent.counterparty),
	}
}

func (man CounterpartyHandshaker) object(parent CounterObject) CounterHandshakeObject {
	prefix := parent.portid + "/channels/" + parent.chanid

	return CounterHandshakeObject{
		CounterObject: man.man.object(parent.portid, parent.chanid),

		State: man.man.protocol.Value([]byte(prefix + "/state")).Enum(),
	}
}

func (man Handshaker) create(ctx sdk.Context, portid, chanid string, channel Channel) (obj HandshakeObject, err error) {
	cobj, err := man.Manager.create(ctx, portid, chanid, channel)
	if err != nil {
		return
	}
	obj = man.object(cobj)

	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, portid, chanid string) (obj HandshakeObject, err error) {
	cobj, err := man.Manager.query(ctx, portid, chanid)
	if err != nil {
		return
	}
	obj = man.object(cobj)

	return obj, nil
}

/*
func (obj HandshakeObject) remove(ctx sdk.Context) {
	obj.Object.remove(ctx)
	obj.State.Delete(ctx)
	obj.counterpartyClient.Delete(ctx)
	obj.NextTimeout.Delete(ctx)
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
	portid, chanid string, channel Channel,
) (HandshakeObject, error) {
	// man.Create() will ensure
	// assert(connectionHops.length === 2)
	// assert(get("channels/{identifier}") === null) and
	// set("channels/{identifier}", connection)
	if len(channel.ConnectionHops) != 1 {
		return HandshakeObject{}, errors.New("ConnectionHops length must be 1")
	}

	obj, err := man.create(ctx, portid, chanid, channel)
	if err != nil {
		return HandshakeObject{}, err
	}

	obj.State.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterparty.{channel,state}
func (man Handshaker) OpenTry(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string, channel Channel,
) (obj HandshakeObject, err error) {
	if len(channel.ConnectionHops) != 1 {
		return HandshakeObject{}, errors.New("ConnectionHops length must be 1")
	}
	obj, err = man.create(ctx, portid, chanid, channel)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return
	}

	if !obj.counterparty.State.Is(ctx, Init) {
		err = errors.New("counterparty state not init")
		return
	}

	if !obj.counterparty.Channel.Is(ctx, Channel{
		Counterparty:     chanid,
		CounterpartyPort: portid,
		ConnectionHops:   []string{obj.Connections[0].GetConnection(ctx).Counterparty},
	}) {
		err = errors.New("wrong counterparty connection")
		return
	}

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.State.Set(ctx, OpenTry)

	return
}

// Using proofs: counterparty.{handshake,state,nextTimeout,clientid,client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string,
) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, portid, chanid)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return
	}

	if !obj.State.Transit(ctx, Init, Open) {
		err = errors.New("ack on non-init connection")
		return
	}

	if !obj.counterparty.Channel.Is(ctx, Channel{
		Counterparty:     chanid,
		CounterpartyPort: portid,
		ConnectionHops:   []string{obj.Connections[0].GetConnection(ctx).Counterparty},
	}) {
		err = errors.New("wrong counterparty")
		return
	}

	if !obj.counterparty.State.Is(ctx, OpenTry) {
		err = errors.New("counterparty state not opentry")
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

	obj.Available.Set(ctx, true)

	return
}

// Using proofs: counterparty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, portid, chanid)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return
	}

	if !obj.State.Transit(ctx, OpenTry, Open) {
		err = errors.New("confirm on non-try connection")
		return
	}

	if !obj.counterparty.State.Is(ctx, Open) {
		err = errors.New("counterparty state not open")
		return
	}

	obj.Available.Set(ctx, true)

	return
}

// TODO
/*
func (obj HandshakeObject) OpenTimeout(ctx sdk.Context) error {
	if !(obj.connection.Client().ConsensusState(ctx).GetHeight()) > obj.NextTimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	switch obj.State.Get(ctx) {
	case Init:
		if !obj.counterparty.connection.Is(ctx, nil) {
			return errors.New("counterparty connection exists")
		}
	case OpenTry:
		if !(obj.counterparty.State.Is(ctx, Init) ||
			obj.counterparty.connection.Is(ctx, nil)) {
			return errors.New("counterparty connection state not init")
		}
		// XXX: check if we need to verify symmetricity for timeout (already proven in OpenTry)
	case Open:
		if obj.counterparty.State.Is(ctx, OpenTry) {
			return errors.New("counterparty connection state not tryopen")
		}
	}

	obj.remove(ctx)

	return nil
}


func (obj HandshakeObject) CloseInit(ctx sdk.Context, nextTimeout uint64) error {
	if !obj.State.Transit(ctx, Open, CloseTry) {
		return errors.New("closeinit on non-open connection")
	}

	obj.NextTimeout.Set(ctx, nextTimeout)

	return nil
}

func (obj HandshakeObject) CloseTry(ctx sdk.Context,  nextTimeoutHeight uint64) error {
	if !obj.State.Transit(ctx, Open, Closed) {
		return errors.New("closetry on non-open connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.State.Is(ctx, CloseTry) {
		return errors.New("unexpected counterparty state value")
	}

	if !obj.counterparty.NextTimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.NextTimeout.Set(ctx, nextTimeoutHeight)

	return nil
}

func (obj HandshakeObject) CloseAck(ctx sdk.Context, timeoutHeight uint64) error {
	if !obj.State.Transit(ctx, CloseTry, Closed) {
		return errors.New("closeack on non-closetry connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if !obj.counterparty.State.Is(ctx, Closed) {
		return errors.New("unexpected counterparty state value")
	}

	if !obj.counterparty.NextTimeout.Is(ctx, timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	obj.NextTimeout.Set(ctx, 0)

	return nil
}

func (obj HandshakeObject) CloseTimeout(ctx sdk.Context) error {
	if !(obj.client.ConsensusState(ctx).GetHeight()) > obj.NextTimeout.Get(ctx)) {
		return errors.New("timeout height not yet reached")
	}

	// XXX: double check if the user can bypass the verification logic somehow
	switch obj.State.Get(ctx) {
	case CloseTry:
		if !obj.counterparty.State.Is(ctx, Open) {
			return errors.New("counterparty connection state not open")
		}
	case Closed:
		if !obj.counterparty.State.Is(ctx, CloseTry) {
			return errors.New("counterparty connection state not closetry")
		}
	}

	obj.State.Set(ctx, Open)
	obj.NextTimeout.Set(ctx, 0)

	return nil

}
*/
