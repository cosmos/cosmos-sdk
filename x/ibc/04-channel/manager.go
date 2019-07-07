package channel

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/state"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Manager is unrestricted
type Manager struct {
	protocol state.Mapping

	connection connection.Manager

	counterparty CounterpartyManager
}

type CounterpartyManager struct {
	protocol commitment.Mapping

	connection connection.CounterpartyManager
}

func NewManager(protocol state.Base, connection connection.Manager) Manager {
	return Manager{
		protocol:     state.NewMapping(protocol, []byte("/connection/")),
		connection:   connection,
		counterparty: NewCounterpartyManager(protocol.Cdc()),
	}
}

func NewCounterpartyManager(cdc *codec.Codec) CounterpartyManager {
	protocol := commitment.NewBase(cdc)

	return CounterpartyManager{
		protocol:   commitment.NewMapping(protocol, []byte("/connection/")),
		connection: connection.NewCounterpartyManager(cdc),
	}
}

// CONTRACT: connection and counterparty must be filled by the caller
func (man Manager) object(connid, chanid string) Object {
	key := connid + "/channels/" + chanid
	return Object{
		chanid:  chanid,
		connid:  connid,
		channel: man.protocol.Value([]byte(key)),

		available: state.NewBoolean(man.protocol.Value([]byte(key + "/available"))),

		seqsend: state.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceSend")), state.Dec),
		seqrecv: state.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceRecv")), state.Dec),
		packets: state.NewIndexer(man.protocol.Prefix([]byte(key+"/packets/")), state.Dec),
	}
}

func (man CounterpartyManager) object(connid, chanid string) CounterObject {
	key := connid + "/channels/" + chanid
	return CounterObject{
		chanid:  chanid,
		connid:  connid,
		channel: man.protocol.Value([]byte(key)),

		available: commitment.NewBoolean(man.protocol.Value([]byte(key + "/available"))),

		seqsend: commitment.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceSend")), state.Dec),
		seqrecv: commitment.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceRecv")), state.Dec),
		packets: commitment.NewIndexer(man.protocol.Prefix([]byte(key+"/packets/")), state.Dec),
	}
}

func (man Manager) create(ctx sdk.Context, connid, chanid string, channel Channel) (obj Object, err error) {
	obj = man.object(connid, chanid)
	if obj.exists(ctx) {
		err = errors.New("channel already exists for the provided id")
		return
	}
	obj.connection, err = man.connection.Query(ctx, connid)
	if err != nil {
		return
	}
	obj.channel.Set(ctx, channel)

	counterconnid := obj.connection.Connection(ctx).Counterparty
	obj.counterparty = man.counterparty.object(counterconnid, channel.Counterparty)
	obj.counterparty.connection = man.counterparty.connection.Object(counterconnid)

	return
}

// Does not check availability
func (man Manager) query(ctx sdk.Context, connid, chanid string) (obj Object, err error) {
	obj = man.object(connid, chanid)
	if !obj.exists(ctx) {
		err = errors.New("channel not exists for the provided id")
		return
	}
	obj.connection, err = man.connection.Query(ctx, connid)
	if err != nil {
		return
	}

	counterconnid := obj.connection.Connection(ctx).Counterparty
	obj.counterparty = man.counterparty.object(counterconnid, obj.Channel(ctx).Counterparty)
	obj.counterparty.connection = man.counterparty.connection.Object(counterconnid)

	return

}

func (man Manager) Query(ctx sdk.Context, connid, chanid string) (obj Object, err error) {
	obj, err = man.query(ctx, connid, chanid)
	if !obj.Available(ctx) {
		err = errors.New("channel not available")
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

func (man PortManager) Create(ctx sdk.Context, connid, chanid string, channel Channel) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	if channel.Port != man.port {
		return Object{}, errors.New("invalid port")
	}

	return man.man.Create(ctx, connid, chanid, channel)
}

func (man PortManager) Query(ctx sdk.Context, connid, chanid string) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	obj, err := man.man.Query(ctx, connid, chanid)
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
	connid string

	protocol state.Mapping
	channel  state.Value

	seqsend state.Integer
	seqrecv state.Integer
	packets state.Indexer

	available state.Boolean

	connection connection.Object

	counterparty CounterObject
}

type CounterObject struct {
	chanid  string
	connid  string
	channel commitment.Value

	seqsend commitment.Integer
	seqrecv commitment.Integer
	packets commitment.Indexer

	available commitment.Boolean

	connection connection.CounterObject
}

func (obj Object) ChanID() string {
	return obj.chanid
}

func (obj Object) Channel(ctx sdk.Context) (res Channel) {
	obj.channel.Get(ctx, &res)
	return
}

func (obj Object) Value(ctx sdk.Context) (res Channel) {
	obj.channel.Get(ctx, &res)
	return
}

func (obj Object) Available(ctx sdk.Context) bool {
	return obj.available.Get(ctx)
}

func (obj Object) Sendable(ctx sdk.Context) bool {
	// TODO: sendable/receivable should be also defined for channels
	return obj.connection.Sendable(ctx)
}

func (obj Object) Receivable(ctx sdk.Context) bool {
	return obj.connection.Receivable(ctx)
}

func (obj Object) SeqSend(ctx sdk.Context) uint64 {
	return obj.seqsend.Get(ctx)
}

func (obj Object) SeqRecv(ctx sdk.Context) uint64 {
	return obj.seqrecv.Get(ctx)
}

func (obj Object) PacketCommit(ctx sdk.Context, index uint64) []byte {
	return obj.packets.Value(index).GetRaw(ctx)
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.channel.Exists(ctx)
}

func (obj Object) Send(ctx sdk.Context, packet Packet) error {
	if !obj.Sendable(ctx) {
		return errors.New("cannot send packets on this channel")
	}

	if uint64(obj.connection.Client().ConsensusState(ctx).GetHeight()) >= packet.Timeout() {
		return errors.New("timeout height higher than the latest known")
	}

	obj.packets.SetRaw(ctx, obj.seqsend.Incr(ctx), packet.Commit())

	return nil
}

func (obj Object) Receive(ctx sdk.Context, packet Packet) error {
	if !obj.Receivable(ctx) {
		return errors.New("cannot receive packets on this channel")
	}

	err := assertTimeout(ctx, packet.Timeout())
	if err != nil {
		return err
	}

	// XXX: increment should happen before verification, reflect on the spec
	if !obj.counterparty.packets.Value(obj.seqrecv.Incr(ctx)).IsRaw(ctx, packet.Commit()) {
		return errors.New("verification failed")
	}

	return nil
}
