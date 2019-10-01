package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type HandshakeStage = byte

const (
	Idle HandshakeStage = iota
	Init
	OpenTry
	Open
	CloseTry
	Closed
)

const HandshakeKind = "handshake"

type Handshaker struct {
	man Manager

	counterParty CounterpartyHandshaker
}

func NewHandshaker(man Manager) Handshaker {
	return Handshaker{
		man: man,
		counterParty: CounterpartyHandshaker{man.counterparty},
	}
}

type CounterpartyHandshaker struct {
	man CounterpartyManager
}

type HandshakeState struct {
	State

	Stage              state.Enum
	CounterpartyClient state.String

	Counterparty CounterHandshakeState
}

type CounterHandshakeState struct {
	CounterState

	Stage              commitment.Enum
	CounterpartyClient commitment.String
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) CreateState(parent State) HandshakeState {
	return HandshakeState{
		State: parent,
		Stage:              man.man.protocol.Value([]byte(parent.id + "/state")).Enum(),
		CounterpartyClient: man.man.protocol.Value([]byte(parent.id + "/counterpartyClient")).String(),

		// CONTRACT: counterParty must be filled by the caller
	}
}

func (man CounterpartyHandshaker) CreateState(id string) CounterHandshakeState {
	return CounterHandshakeState{
		CounterState: man.man.CreateState(id),
		Stage:              man.man.protocol.Value([]byte(id + "/state")).Enum(),
		CounterpartyClient: man.man.protocol.Value([]byte(id + "/counterpartyClient")).String(),
	}
}

func (man Handshaker) create(ctx sdk.Context, id string, connection Connection, counterpartyClient string) (obj HandshakeState, err error) {
	cobj, err := man.man.create(ctx, id, connection, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.CreateState(cobj)
	obj.CounterpartyClient.Set(ctx, counterpartyClient)
	obj.Counterparty = man.counterParty.CreateState(connection.Counterparty)
	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, id string) (obj HandshakeState, err error) {
	cobj, err := man.man.query(ctx, id, HandshakeKind)
	if err != nil {
		return
	}
	obj = man.CreateState(cobj)
	obj.Counterparty = man.counterParty.CreateState(obj.GetConnection(ctx).Counterparty)
	return
}

func (obj HandshakeState) remove(ctx sdk.Context) {
	obj.State.remove(ctx)
	obj.Stage.Delete(ctx)
	obj.CounterpartyClient.Delete(ctx)
}

// Using proofs: none
func (man Handshaker) OpenInit(ctx sdk.Context,
	id string, connection Connection, counterpartyClient string,
) (HandshakeState, error) {
	// man.Create() will ensure
	// assert(get("connections/{identifier}") === null) and
	// set("connections{identifier}", connection)
	obj, err := man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return HandshakeState{}, err
	}

	obj.Stage.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterParty.{connection,state,nextTimeout,counterpartyClient, client}
func (man Handshaker) OpenTry(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	id string, connection Connection, counterpartyClient string,
) (obj HandshakeState, err error) {
	obj, err = man.create(ctx, id, connection, counterpartyClient)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, height, proofs)
	if err != nil {
		return
	}

	if !obj.Counterparty.Stage.Is(ctx, Init) {
		err = errors.New("counterParty state not init")
		return
	}

	if !obj.Counterparty.Connection.Is(ctx, Connection{
		Client:       counterpartyClient,
		Counterparty: id,
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterParty connection")
		return
	}

	if !obj.Counterparty.CounterpartyClient.Is(ctx, connection.Client) {
		err = errors.New("counterParty client not match")
		return
	}

	// TODO: commented out, need to check whether the stored client is compatible
	// make a separate module that manages recent n block headers
	// ref #4647
	/*
		var expected client.ConsensusState
		obj.self.Get(ctx, expheight, &expected)
		if !obj.counterParty.client.Is(ctx, expected) {
			return errors.New("unexpected counterParty client value")
		}
	*/

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.Stage.Set(ctx, OpenTry)

	return
}

// Using proofs: counterParty.{connection, state, timeout, counterpartyClient, client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	id string,
) (obj HandshakeState, err error) {
	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, height, proofs)
	if err != nil {
		return
	}

	if !obj.Stage.Transit(ctx, Init, Open) {
		err = errors.New("ack on non-init connection")
		return
	}

	if !obj.Counterparty.Connection.Is(ctx, Connection{
		Client:       obj.CounterpartyClient.Get(ctx),
		Counterparty: obj.ID(),
		Path:         obj.path,
	}) {
		err = errors.New("wrong counterParty")
		return
	}

	if !obj.Counterparty.Stage.Is(ctx, OpenTry) {
		err = errors.New("counterParty state not opentry")
		return
	}

	if !obj.Counterparty.CounterpartyClient.Is(ctx, obj.GetConnection(ctx).Client) {
		err = errors.New("counterParty client not match")
		return
	}

	// TODO: implement in v1
	/*
		var expected client.ConsensusState
		// obj.self.Get(ctx, expheight, &expected)
		if !obj.counterParty.client.Is(ctx, expected) {
			// return errors.New("unexpected counterParty client value")
		}
	*/
	obj.Available.Set(ctx, true)

	return
}

// Using proofs: counterParty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	id string) (obj HandshakeState, err error) {

	obj, err = man.query(ctx, id)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, height, proofs)
	if err != nil {
		return
	}

	if !obj.Stage.Transit(ctx, OpenTry, Open) {
		err = errors.New("confirm on non-try connection")
		return
	}

	if !obj.Counterparty.Stage.Is(ctx, Open) {
		err = errors.New("counterParty state not open")
		return
	}

	obj.Available.Set(ctx, true)

	return
}
