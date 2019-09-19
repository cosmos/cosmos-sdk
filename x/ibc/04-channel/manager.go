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

	counterparty CounterpartyManager

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
		counterparty: NewCounterpartyManager(protocol.Cdc()),

		ports: make(map[string]struct{}),
	}
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	protocol := commitment.NewMapping(cdc, nil)

	return CounterpartyManager{
		protocol:   protocol.Prefix(LocalRoot()),
		connection: connection.NewCounterpartyManager(cdc),
	}
}

// CONTRACT: connection and counterparty must be filled by the caller
func (man Manager) object(portid, chanid string) Object {
	key := portid + "/channels/" + chanid
	return Object{
		chanid:  chanid,
		portid:  portid,
		Channel: man.protocol.Value([]byte(key)),

		Available: man.protocol.Value([]byte(key + "/available")).Boolean(),

		SeqSend: man.protocol.Value([]byte(key + "/nextSequenceSend")).Integer(state.Dec),
		SeqRecv: man.protocol.Value([]byte(key + "/nextSequenceRecv")).Integer(state.Dec),
		Packets: man.protocol.Prefix([]byte(key + "/packets/")).Indexer(state.Dec),
	}
}

func (man CounterpartyManager) object(portid, chanid string) CounterObject {
	key := portid + "/channels/" + chanid
	return CounterObject{
		chanid:  chanid,
		portid:  portid,
		Channel: man.protocol.Value([]byte(key)),

		Available: man.protocol.Value([]byte(key + "/available")).Boolean(),

		SeqSend: man.protocol.Value([]byte(key + "/nextSequenceSend")).Integer(state.Dec),
		SeqRecv: man.protocol.Value([]byte(key + "/nextSequenceRecv")).Integer(state.Dec),
		Packets: man.protocol.Prefix([]byte(key + "/packets/")).Indexer(state.Dec),
	}
}

func (man Manager) create(ctx sdk.Context, portid, chanid string, channel Channel) (obj Object, err error) {
	obj = man.object(portid, chanid)
	if obj.exists(ctx) {
		err = errors.New("channel already exists for the provided id")
		return
	}
	obj.Channel.Set(ctx, channel)
	obj.counterparty = man.counterparty.object(channel.CounterpartyPort, channel.Counterparty)

	for _, hop := range channel.ConnectionHops {
		connobj, err := man.connection.Query(ctx, hop)
		if err != nil {
			return obj, err
		}
		obj.Connections = append(obj.Connections, connobj)
	}

	for _, hop := range channel.CounterpartyHops() {
		connobj := man.counterparty.connection.Object(hop)
		obj.counterparty.Connections = append(obj.counterparty.Connections, connobj)
	}

	return
}

// Does not check availability
func (man Manager) query(ctx sdk.Context, portid, chanid string) (obj Object, err error) {
	obj = man.object(portid, chanid)
	if !obj.exists(ctx) {
		err = errors.New("channel not exists for the provided id")
		return
	}

	channel := obj.GetChannel(ctx)
	obj.counterparty = man.counterparty.object(channel.CounterpartyPort, channel.Counterparty)
	for _, hop := range channel.ConnectionHops {
		connobj, err := man.connection.Query(ctx, hop)
		if err != nil {
			return obj, err
		}
		obj.Connections = append(obj.Connections, connobj)
	}

	for _, hop := range channel.CounterpartyHops() {
		connobj := man.counterparty.connection.Object(hop)
		obj.counterparty.Connections = append(obj.counterparty.Connections, connobj)
	}

	return

}

func (man Manager) Query(ctx sdk.Context, portid, chanid string) (obj Object, err error) {
	obj, err = man.query(ctx, portid, chanid)
	if !obj.Available.Get(ctx) {
		err = errors.New("channel not Available")
		return
	}
	return
}

// TODO
/*
func (man Manager) Port(port string, chanid func(string) bool) PortManager {
	return PortManager{
		man:    man,
		port:   le,
		chanid: chanid,
	}
}

// PortManage is port specific
type PortManager struct {
	man    Manager
	port   string
	chanid func(string) bool
}

func (man PortManager) Create(ctx sdk.Context, portid, chanid string, channel Channel) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	if channel.Port != man.port {
		return Object{}, errors.New("invalid port")
	}

	return man.man.Create(ctx, portid, chanid, channel)
}

func (man PortManager) Query(ctx sdk.Context, portid, chanid string) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	obj, err := man.man.Query(ctx, portid, chanid)
	if err != nil {
		return Object{}, err
	}

	if obj.Value(ctx).Port != man.port {
		return Object{}, errors.New("invalid port")
	}

	return obj, nil
}
*/

type Object struct {
	chanid string
	portid string

	Channel state.Value

	SeqSend state.Integer
	SeqRecv state.Integer
	Packets state.Indexer

	Available state.Boolean

	Connections []connection.Object

	counterparty CounterObject
}

type CounterObject struct {
	chanid string
	portid string

	Channel commitment.Value

	SeqSend commitment.Integer
	SeqRecv commitment.Integer
	Packets commitment.Indexer

	Available commitment.Boolean

	Connections []connection.CounterObject
}

func (obj Object) OriginConnection() connection.Object {
	return obj.Connections[0]
}

func (obj Object) Context(ctx sdk.Context, proofs []commitment.Proof, height uint64) (sdk.Context, error) {
	return obj.OriginConnection().Context(ctx, height, proofs)
}

func (obj Object) ChanID() string {
	return obj.chanid
}

func (obj Object) GetChannel(ctx sdk.Context) (res Channel) {
	obj.Channel.Get(ctx, &res)
	return
}

func (obj Object) PacketCommit(ctx sdk.Context, index uint64) []byte {
	return obj.Packets.Value(index).GetRaw(ctx)
}

/*
func (obj Object) Sendable(ctx sdk.Context) bool {
	return obj.connection
}

func (obj Object) Receivable(ctx sdk.Context) bool {
	return kinds[obj.kind.Get(ctx)].Receivable
}
*/
func (obj Object) exists(ctx sdk.Context) bool {
	return obj.Channel.Exists(ctx)
}

func (man Manager) Send(ctx sdk.Context, portid, chanid string, packet Packet) error {
	/*
		if !obj.Sendable(ctx) {
			return errors.New("cannot send Packets on this channel")
		}
	*/
	if portid != packet.SenderPort() {
		return errors.New("Invalid portid")
	}

	obj, err := man.Query(ctx, portid, chanid)
	if err != nil {
		return err
	}

	if obj.OriginConnection().Client.GetConsensusState(ctx).GetHeight() >= packet.Timeout() {
		return errors.New("timeout height higher than the latest known")
	}

	obj.Packets.Set(ctx, obj.SeqSend.Increment(ctx), packet)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			EventTypeSendPacket,
			sdk.NewAttribute(AttributeKeySenderPort, packet.SenderPort()),
			sdk.NewAttribute(AttributeKeyReceiverPort, packet.ReceiverPort()),
			sdk.NewAttribute(AttributeKeyChannelID, chanid),
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

	// XXX: increment should happen before verification, reflect on the spec
	// TODO: packet should be custom marshalled
	if !obj.counterparty.Packets.Value(obj.SeqRecv.Increment(ctx)).Is(ctx, packet) {
		return errors.New("verification failed")
	}

	return nil
}
