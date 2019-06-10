package channel

import (
	"github.com/cosmos/cosmos-sdk/store/mapping"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type Manager struct {
	protocol mapping.Mapping

	connection connection.Manager

	remote *Manager
}

func NewManager(protocol mapping.Base, connection connection.Manager) Manager {
	return Manager{
		protocol:   mapping.NewMapping(protocol, []byte("/")),
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
	obj := man.object(connid, chanid)
	if obj.exists(ctx) {
		err = errors.New("channel already exists for the provided id")
		return
	}
	obj.connection, err = man.connection.Query(ctx, connid)
	if err != nil {
		return
	}
	obj.channel.Set(ctx, channel)
	remote := man.remote.object()
}

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

func (obj Object) OpenInit(ctx sdk.Context) error {
	// OpenInit will
}
