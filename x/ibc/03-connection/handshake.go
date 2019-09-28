package connection

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

	State              state.Enum
	CounterpartyClient state.String

	Counterparty CounterHandshakeObject
}

type CounterHandshakeObject struct {
	CounterObject

	State              commitment.Enum
	CounterpartyClient commitment.String
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) Object(parent Object) HandshakeObject {
	return HandshakeObject{
		Object: parent,

		State:              man.man.protocol.Value([]byte(parent.id + "/state")).Enum(),
		CounterpartyClient: man.man.protocol.Value([]byte(parent.id + "/counterpartyClient")).String(),

		// CONTRACT: counterparty must be filled by the caller
	}
}

func (man CounterpartyHandshaker) Object(id string) CounterHandshakeObject {
	return CounterHandshakeObject{
		CounterObject: man.man.Object(id),

		State:              man.man.protocol.Value([]byte(id + "/state")).Enum(),
		CounterpartyClient: man.man.protocol.Value([]byte(id + "/counterpartyClient")).String(),
	}
}

func (man Handshaker) create(ctx sdk.Context, id string, connection Connection, counterpartyClient string) (obj HandshakeObject, err error) {
	cobj, err := man.man.create(ctx, id, connection, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.Object(cobj)
	obj.CounterpartyClient.Set(ctx, counterpartyClient)
	obj.Counterparty = man.counterparty.Object(connection.Counterparty)
	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, id string) (obj HandshakeObject, err error) {
	cobj, err := man.man.query(ctx, id, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.Object(cobj)
	obj.Counterparty = man.counterparty.Object(obj.GetConnection(ctx).Counterparty)
	return
}

// Using proofs: none
func (man Handshaker) OpenInit(ctx sdk.Context,
	id string, connection Connection, counterpartyClient string,
) (HandshakeObject, error) {
	// man.Create() will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)
	obj, err := man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return HandshakeObject{}, err
	}

	obj.State.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterparty.{connection,state,nextTimeout,counterpartyClient, client}
func (man Handshaker) OpenTry(ctx sdk.Context,
	proofs []commitment.Proof,
	id string, connection Connection, counterpartyClient string,
) (obj HandshakeObject, err error) {
	obj, err = man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, connection.Path, proofs)
	if err != nil {
		return
	}

	if !obj.Counterparty.State.Is(ctx, Init) {
		err = errors.New("counterparty state not init")
		return
	}

	if !obj.Counterparty.Connection.Is(ctx, Connection{
		Client:       counterpartyClient,
		Counterparty: id,
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterparty connection")
		return
	}

	if !obj.Counterparty.CounterpartyClient.Is(ctx, connection.Client) {
		err = errors.New("counterparty client not match")
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

	obj.State.Set(ctx, OpenTry)

	return
}

// Using proofs: counterparty.{connection, state, counterpartyClient, client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	proofs []commitment.Proof,
	id string, /*expheight uint64, */
) (obj HandshakeObject, err error) {
	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, nil, proofs)
	if err != nil {
		return
	}

	if !obj.State.Transit(ctx, Init, Open) {
		err = errors.New("ack on non-init connection")
		return
	}

	if !obj.Counterparty.Connection.Is(ctx, Connection{
		Client:       obj.CounterpartyClient.Get(ctx),
		Counterparty: obj.ID(),
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterparty")
		return
	}

	if !obj.Counterparty.State.Is(ctx, OpenTry) {
		err = errors.New("counterparty state not opentry")
		return
	}

	if !obj.Counterparty.CounterpartyClient.Is(ctx, obj.GetConnection(ctx).Client) {
		err = errors.New("counterparty client not match")
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
	obj.Available.Set(ctx, true)

	return
}

// Using proofs: counterparty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context,
	proofs []commitment.Proof,
	id string) (obj HandshakeObject, err error) {

	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, nil, proofs)
	if err != nil {
		return
	}

	if !obj.State.Transit(ctx, OpenTry, Open) {
		err = errors.New("confirm on non-try connection")
		return
	}

	if !obj.Counterparty.State.Is(ctx, Open) {
		err = errors.New("counterparty state not open")
		return
	}

	obj.Available.Set(ctx, true)

	return
}
