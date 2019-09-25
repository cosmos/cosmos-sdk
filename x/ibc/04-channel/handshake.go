package channel

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

type Handshaker struct {
	Manager

	counterParty CounterpartyHandshaker
}

func (man Handshaker) Kind() string {
	return "handshake"
}

func NewHandshaker(man Manager) Handshaker {
	return Handshaker{
		Manager:      man,
		counterParty: CounterpartyHandshaker{man.counterParty},
	}
}

type CounterpartyHandshaker struct {
	man CounterpartyManager
}

type HandshakeState struct {
	State

	HandshakeStage state.Enum

	counterParty CounterHandshakeState
}

type CounterHandshakeState struct {
	CounterState

	Stage commitment.Enum
}

// CONTRACT: client and remote must be filled by the caller
func (man Handshaker) createState(parent State) HandshakeState {
	prefix := parent.portId + "/channels/" + parent.chanId

	return HandshakeState{
		State:          parent,
		HandshakeStage: man.protocol.Value([]byte(prefix + "/state")).Enum(),
		counterParty:   man.counterParty.createState(parent.counterParty),
	}
}

func (man CounterpartyHandshaker) createState(parent CounterState) CounterHandshakeState {
	prefix := parent.portId + "/channels/" + parent.chanId

	return CounterHandshakeState{
		CounterState: man.man.object(parent.portId, parent.chanId),
		Stage:        man.man.protocol.Value([]byte(prefix + "/state")).Enum(),
	}
}

func (man Handshaker) create(ctx sdk.Context, portid, chanid string, channel Channel) (obj HandshakeState, err error) {
	cobj, err := man.Manager.create(ctx, portid, chanid, channel)
	if err != nil {
		return
	}
	obj = man.createState(cobj)

	return obj, nil
}

func (man Handshaker) query(ctx sdk.Context, portid, chanid string) (obj HandshakeState, err error) {
	cobj, err := man.Manager.query(ctx, portid, chanid)
	if err != nil {
		return
	}
	obj = man.createState(cobj)

	return obj, nil
}

/*
func (obj HandshakeState) remove(ctx sdk.Context) {
	obj.Stage.remove(ctx)
	obj.HandshakeStage.Delete(ctx)
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
) (HandshakeState, error) {
	// man.Create() will ensure
	// assert(connectionHops.length === 2)
	// assert(get("channels/{identifier}") === null) and
	// set("channels/{identifier}", connection)
	if len(channel.ConnectionHops) != 1 {
		return HandshakeState{}, errors.New("ConnectionHops length must be 1")
	}

	obj, err := man.create(ctx, portid, chanid, channel)
	if err != nil {
		return HandshakeState{}, err
	}

	obj.HandshakeStage.Set(ctx, Init)

	return obj, nil
}

// Using proofs: counterParty.{channel,state}
func (man Handshaker) OpenTry(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string, channel Channel,
) (obj HandshakeState, err error) {
	if len(channel.ConnectionHops) != 1 {
		return HandshakeState{}, errors.New("ConnectionHops length must be 1")
	}
	obj, err = man.create(ctx, portid, chanid, channel)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return
	}

	if !obj.counterParty.Stage.Is(ctx, Init) {
		err = errors.New("counterParty state not init")
		return
	}

	if !obj.counterParty.Channel.Is(ctx, Channel{
		Counterparty:     chanid,
		CounterpartyPort: portid,
		ConnectionHops:   []string{obj.Connections[0].GetConnection(ctx).Counterparty},
	}) {
		err = errors.New("wrong counterParty connection")
		return
	}

	// CONTRACT: OpenTry() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get("connections/{desiredIdentifier}") === null) and
	// set("connections{identifier}", connection)

	obj.HandshakeStage.Set(ctx, OpenTry)

	return
}

// Using proofs: counterParty.{handshake,state,nextTimeout,clientid,client}
func (man Handshaker) OpenAck(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string,
) (obj HandshakeState, err error) {
	obj, err = man.query(ctx, portid, chanid)
	if err != nil {
		return
	}

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return
	}

	if !obj.HandshakeStage.Transit(ctx, Init, Open) {
		err = errors.New("ack on non-init connection")
		return
	}

	if !obj.counterParty.Channel.Is(ctx, Channel{
		Counterparty:     chanid,
		CounterpartyPort: portid,
		ConnectionHops:   []string{obj.Connections[0].GetConnection(ctx).Counterparty},
	}) {
		err = errors.New("wrong counterParty")
		return
	}

	if !obj.counterParty.Stage.Is(ctx, OpenTry) {
		err = errors.New("counterParty state not opentry")
		return
	}

	// TODO: commented out, implement in v1
	/*
		var expected client.ConsensusState
		obj.self.Get(ctx, expheight, &expected)
		if !obj.counterParty.client.Is(ctx, expected) {
			return errors.New("unexpected counterParty client value")
		}
	*/

	obj.Available.Set(ctx, true)

	return
}

// Using proofs: counterParty.{connection,state, nextTimeout}
func (man Handshaker) OpenConfirm(ctx sdk.Context,
	proofs []commitment.Proof, height uint64,
	portid, chanid string) (obj HandshakeState, err error) {
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

	if !obj.counterParty.Stage.Is(ctx, Open) {
		err = errors.New("counterParty state not open")
		return
	}

	obj.Available.Set(ctx, true)

	return
}

// TODO
/*
func (obj HandshakeState) CloseInit(ctx sdk.Context, nextTimeout uint64) error {
	if !obj.HandshakeStage.Transit(ctx, Open, CloseTry) {
		return errors.New("closeinit on non-open connection")
	}

	obj.NextTimeout.Set(ctx, nextTimeout)

	return nil
}
*/
