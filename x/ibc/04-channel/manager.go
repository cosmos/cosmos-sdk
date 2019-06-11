package channel

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

// Manager is unrestricted
type Manager struct {
	protocol mapping.Mapping

	connection connection.Manager

	remote *Manager
}

func NewManager(protocol mapping.Base, connection connection.Manager) Manager {
	man := newManager(protocol, connection)
	man.remote = newRemoteManager(protocol.Cdc())
	return man
}

func newRemoteManager(cdc *codec.Codec) *Manager {
	protocol := mapping.NewBaseWithGetter(cdc, commitment.GetStore)
	connection := connection.NewRemoteManager(cdc)
	res := newManager(protocol, *connection)
	return &res
}

func newManager(protocol mapping.Base, connection connection.Manager) Manager {
	return Manager{
		protocol:   mapping.NewMapping(protocol, []byte("/connection/")),
		connection: connection,
	}
}

// CONTRACT: remote must be filled by the caller
func (man Manager) object(connid, chanid string) Object {
	key := connid + "/channels/" + chanid
	return Object{
		connid:      connid,
		chanid:      chanid,
		channel:     man.protocol.Value([]byte(key)),
		state:       mapping.NewEnum(man.protocol.Value([]byte(key + "/state"))),
		nexttimeout: mapping.NewInteger(man.protocol.Value([]byte(key+"/timeout")), mapping.Dec),

		// TODO: remove length functionality from mapping.Indeer(will be handled manually)
		seqsend: mapping.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceSend")), mapping.Dec),
		seqrecv: mapping.NewInteger(man.protocol.Value([]byte(key+"/nextSequenceRecv")), mapping.Dec),
		packets: mapping.NewIndexer(man.protocol.Prefix([]byte(key+"/packets/")), mapping.Dec),
	}
}

func (man Manager) Create(ctx sdk.Context, connid, chanid string, channel Channel) (obj Object, err error) {
	obj = man.object(connid, chanid)
	if obj.exists(ctx) {
		err = errors.New("channel already exists for the provided id")
		return
	}
	obj.connection, err = man.connection.Query(ctx, connid)
	if err != nil {
		return
	}
	if obj.connection.State(ctx) != connection.Open {
		err = errors.New("connection exists but not opened")
		return
	}
	obj.channel.Set(ctx, channel)
	remote := man.remote.object(obj.connection.Value(ctx).Counterparty, channel.Counterparty)
	obj.remote = &remote
	return
}

func (man Manager) Query(ctx sdk.Context, connid, chanid string) (obj Object, err error) {
	obj = man.object(connid, chanid)
	if !obj.exists(ctx) {
		err = errors.New("channel not exists for the provided id")
		return
	}
	obj.connection, err = man.connection.Query(ctx, connid)
	if err != nil {
		return
	}
	if obj.connection.State(ctx) != connection.Open {
		err = errors.New("connection exists but not opened")
		return
	}

	channel := obj.Value(ctx)
	remote := man.remote.object(obj.connection.Value(ctx).Counterparty, channel.Counterparty)
	obj.remote = &remote
	return
}

func (man Manager) Module(module string, chanid func(string) bool) ModuleManager {
	return ModuleManager{
		man:    man,
		module: module,
		chanid: chanid,
	}
}

// ModuleManage is module specific
type ModuleManager struct {
	man    Manager
	module string
	chanid func(string) bool
}

func (man ModuleManager) Create(ctx sdk.Context, connid, chanid string, channel Channel) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	if channel.Module != man.module {
		return Object{}, errors.New("invalid module")
	}

	return man.man.Create(ctx, connid, chanid, channel)
}

func (man ModuleManager) Query(ctx sdk.Context, connid, chanid string) (Object, error) {
	if !man.chanid(chanid) {
		return Object{}, errors.New("invalid channel id")
	}

	obj, err := man.man.Query(ctx, connid, chanid)
	if err != nil {
		return Object{}, err
	}

	if obj.Value(ctx).Module != man.module {
		return Object{}, errors.New("invalid module")
	}

	return obj, nil
}

// XXX: remove connid(already exists in connection.id)
type Object struct {
	connid      string
	chanid      string
	channel     mapping.Value
	state       mapping.Enum
	nexttimeout mapping.Integer

	seqsend mapping.Integer
	seqrecv mapping.Integer
	packets mapping.Indexer

	connection connection.Object

	// CONTRACT: remote should not be used when remote
	remote *Object
}

func (obj Object) ConnID() string {
	return obj.connid
}

func (obj Object) ChanID() string {
	return obj.chanid
}

func (obj Object) Value(ctx sdk.Context) (res Channel) {
	obj.channel.Get(ctx, &res)
	return
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.channel.Exists(ctx)
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if uint64(ctx.BlockHeight()) > timeoutHeight {
		return errors.New("timeout")
	}
	return nil
}

// TODO: ocapify callingModule
func (obj Object) OpenInit(ctx sdk.Context) error {
	// CONTRACT: OpenInit() should be called after man.Create(), not man.Query(),
	// which will ensure
	// assert(get() === null) and
	// set() and
	// connection.state == open

	// getCallingModule() === channel.moduleIdentifier is ensured by ModuleManager

	if !obj.state.Transit(ctx, Idle, Init) {
		return errors.New("init on non-idle channel")
	}

	obj.seqsend.Set(ctx, 0)
	obj.seqrecv.Set(ctx, 0)

	return nil
}

func (obj Object) OpenTry(ctx sdk.Context, timeoutHeight, nextTimeoutHeight uint64) error {
	if !obj.state.Transit(ctx, Idle, OpenTry) {
		return errors.New("opentry on non-idle channel")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	// XXX
}

func (obj Object) Send(ctx sdk.Context, packet Packet) error {
	if obj.state.Get(ctx) != Open {
		return errors.New("send on non-open channel")
	}

	if obj.connection.State(ctx) != Open {
		return errors.New("send on non-open connection")
	}

	if uint64(obj.connection.Client(ctx).GetBase().GetHeight()) >= packet.Timeout() {
		return errors.New("timeout height higher than the latest known")
	}

	obj.packets.Set(ctx, obj.seqsend.Incr(ctx), packet)

	return nil
}

func (obj Object) Receive(ctx sdk.Context, packet Packet) error {
	if obj.state.Get(ctx) != Open {
		return errors.New("send on non-open channel")
	}

	if obj.connection.State(ctx) != Open {
		return errors.New("send on non-open connection")
	}

	err := assertTimeout(ctx, packet.Timeout())
	if err != nil {
		return err
	}

	// XXX: increment should happen before verification, reflect on the spec
	if !obj.remote.packets.Value(obj.seqrecv.Incr(ctx)).Is(ctx, packet.Commit()) {
		return errors.New("verification failed")
	}

	return nil
}
