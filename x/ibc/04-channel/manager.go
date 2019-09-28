package channel

import (
	"errors"
	"strconv"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Manager is unrestricted
type Manager struct {
	protocol state.Mapping

	connection connection.Manager

	counterParty CounterpartyManager

	ports map[string]struct{}
}

type CounterpartyManager struct {
	protocol commitment.Mapping

	connection connection.CounterpartyManager
}

func NewManager(protocol state.Mapping, connection connection.Manager) Manager {
	return Manager{
		protocol:     protocol.Prefix(LocalRoot()),
		connection:   connection,
		counterParty: NewCounterpartyManager(protocol.Cdc()),
		ports:        make(map[string]struct{}),
	}
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	protocol := commitment.NewMapping(cdc, nil)

	return CounterpartyManager{
		protocol:   protocol.Prefix(LocalRoot()),
		connection: connection.NewCounterpartyManager(cdc),
	}
}

// CONTRACT: connection and counterParty must be filled by the caller
func (man Manager) object(portId, chanId string) State {
	key := portId + "/channels/" + chanId
	return State{
		chanId:    chanId,
		portId:    portId,
		Channel:   man.protocol.Value([]byte(key)),
		Available: man.protocol.Value([]byte(key + "/available")).Boolean(),
		SeqSend:   man.protocol.Value([]byte(key + "/nextSequenceSend")).Integer(state.Dec),
		SeqRecv:   man.protocol.Value([]byte(key + "/nextSequenceRecv")).Integer(state.Dec),
		Packets:   man.protocol.Prefix([]byte(key + "/packets/")).Indexer(state.Dec),
	}
}

func (man CounterpartyManager) object(portid, chanid string) CounterState {
	key := portid + "/channels/" + chanid
	return CounterState{
		chanId:    chanid,
		portId:    portid,
		Channel:   man.protocol.Value([]byte(key)),
		Available: man.protocol.Value([]byte(key + "/available")).Boolean(),
		SeqSend:   man.protocol.Value([]byte(key + "/nextSequenceSend")).Integer(state.Dec),
		SeqRecv:   man.protocol.Value([]byte(key + "/nextSequenceRecv")).Integer(state.Dec),
		Packets:   man.protocol.Prefix([]byte(key + "/packets/")).Indexer(state.Dec),
	}
}

func (man Manager) create(ctx sdk.Context, portid, chanid string, channel Channel) (obj State, err error) {
	obj = man.object(portid, chanid)
	if obj.exists(ctx) {
		err = errors.New("channel already exists for the provided id")
		return
	}
	obj.Channel.Set(ctx, channel)
	obj.counterParty = man.counterParty.object(channel.CounterpartyPort, channel.Counterparty)

	for _, hop := range channel.ConnectionHops {
		connobj, err := man.connection.Query(ctx, hop)
		if err != nil {
			return obj, err
		}
		obj.Connections = append(obj.Connections, connobj)
	}

	for _, hop := range channel.CounterpartyHops() {
		connobj := man.counterParty.connection.CreateState(hop)
		obj.counterParty.Connections = append(obj.counterParty.Connections, connobj)
	}

	return
}

// Does not check availability
func (man Manager) query(ctx sdk.Context, portid, chanid string) (obj State, err error) {
	obj = man.object(portid, chanid)
	if !obj.exists(ctx) {
		err = errors.New("channel not exists for the provided id")
		return
	}

	channel := obj.GetChannel(ctx)
	obj.counterParty = man.counterParty.object(channel.CounterpartyPort, channel.Counterparty)
	for _, hop := range channel.ConnectionHops {
		connobj, err := man.connection.Query(ctx, hop)
		if err != nil {
			return obj, err
		}
		obj.Connections = append(obj.Connections, connobj)
	}

	for _, hop := range channel.CounterpartyHops() {
		connobj := man.counterParty.connection.CreateState(hop)
		obj.counterParty.Connections = append(obj.counterParty.Connections, connobj)
	}

	return

}

func (man Manager) Query(ctx sdk.Context, portid, chanid string) (obj State, err error) {
	obj, err = man.query(ctx, portid, chanid)
	if !obj.Available.Get(ctx) {
		err = errors.New("channel not available")
		return
	}
	return
}

type State struct {
	chanId string
	portId string

	Channel state.Value

	SeqSend state.Integer
	SeqRecv state.Integer
	Packets state.Indexer

	Available state.Boolean

	Connections []connection.State

	counterParty CounterState
}

type CounterState struct {
	chanId string
	portId string

	Channel commitment.Value

	SeqSend commitment.Integer
	SeqRecv commitment.Integer
	Packets commitment.Indexer

	Available commitment.Boolean

	Connections []connection.CounterState
}

func (obj State) OriginConnection() connection.State {
	return obj.Connections[0]
}

func (obj State) Context(ctx sdk.Context, proofs []commitment.Proof, height uint64) (sdk.Context, error) {
	return obj.OriginConnection().Context(ctx, height, proofs)
}

func (obj State) ChanID() string {
	return obj.chanId
}

func (obj State) GetChannel(ctx sdk.Context) (res Channel) {
	obj.Channel.Get(ctx, &res)
	return
}

func (obj State) PacketCommit(ctx sdk.Context, index uint64) []byte {
	return obj.Packets.Value(index).GetRaw(ctx)
}

/*
func (obj Stage) Sendable(ctx sdk.Context) bool {
	return obj.connection
}

func (obj Stage) Receivable(ctx sdk.Context) bool {
	return kinds[obj.kind.Get(ctx)].Receivable
}
*/
func (obj State) exists(ctx sdk.Context) bool {
	return obj.Channel.Exists(ctx)
}

func (man Manager) Send(ctx sdk.Context, chanId string, packet Packet) error {
	obj, err := man.Query(ctx, packet.SenderPort(), chanId)
	if err != nil {
		return err
	}

	if obj.OriginConnection().Client.GetConsensusState(ctx).GetHeight() >= packet.Timeout() {
		return errors.New("timeout height higher than the latest known")
	}

	obj.Packets.SetRaw(ctx, obj.SeqSend.Increment(ctx), packet.Marshal())

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeSendPacket,
			sdk.NewAttribute(AttributeKeySenderPort, packet.SenderPort()),
			sdk.NewAttribute(AttributeKeyReceiverPort, packet.ReceiverPort()),
			sdk.NewAttribute(AttributeKeyChannelID, chanId),
			sdk.NewAttribute(AttributeKeySequence, strconv.FormatUint(obj.SeqSend.Get(ctx), 10)),
		),
	})

	return nil
}

func (man Manager) Receive(ctx sdk.Context, proofs []commitment.Proof, height uint64, portid, chanid string, packet Packet) error {
	obj, err := man.Query(ctx, portid, chanid)
	if err != nil {
		return err
	}

	/*
		if !obj.Receivable(ctx) {
			return errors.New("cannot receive Packets on this channel")
		}
	*/

	ctx, err = obj.Context(ctx, proofs, height)
	if err != nil {
		return err
	}

	err = assertTimeout(ctx, packet.Timeout())
	if err != nil {
		return err
	}

	if !obj.counterParty.Packets.Value(obj.SeqRecv.Increment(ctx)).IsRaw(ctx, packet.Marshal()) {
		return errors.New("verification failed")
	}

	return nil
}
