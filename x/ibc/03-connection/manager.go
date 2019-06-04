package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

type HandshakeManager struct {
	protocol mapping.Mapping

	self mapping.Indexer

	client client.Manager
}

func NewHandshakeManager(protocol, free mapping.Base, client client.Manager) HandshakeManager {
	return HandshakeManager{
		protocol: mapping.NewMapping(protocol, []byte("/")),

		self: mapping.NewIndexer(free, []byte("/self"), mapping.Dec),

		client: client,
	}
}

func (man HandshakeManager) object(id string) Object {
	return Object{
		id:         id,
		connection: man.protocol.Value([]byte(id)),
		state:      man.protocol.Value([]byte(id + "/state")).Enum(),

		self: man.self,
	}
}

func (man HandshakeManager) Query(ctx sdk.Context, key string) (res Object, st State) {
	res = man.object(key)
	st = res.state.Get(ctx)
	return
}

type Object struct {
	id           string
	counterparty string
	connection   mapping.Value
	state        mapping.Enum
	nexttimeout  mapping.Integer
	client       client.Object

	self mapping.Indexer
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint64) error {
	if ctx.BlockHeight() > int64(timeoutHeight) {
		return errors.New("timeout")
	}

	return nil
}

func (obj Object) Value(ctx sdk.Context) (res Connection) {
	obj.connection.Get(ctx, &res)
	return
}

func (obj Object) OpenInit(ctx sdk.Context, desiredCounterparty, client, counterpartyClient string, nextTimeoutHeight uint64) error {
	if obj.exists(ctx) {
		panic("init on existing connection")
	}

	if !obj.state.Transit(ctx, Idle, Init) {
		panic("init on non-idle connection")
	}

	obj.connection.Set(ctx, Connection{
		Counterparty:       desiredCounterparty,
		Client:             client,
		CounterpartyClient: counterpartyClient,
	})
	obj.nexttimeout.Set(ctx, int64(nextTimeoutHeight))

	return nil
}

func (obj Object) OpenAck(ctx sdk.Context, remote Object, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {
	if obj.counterparty != remote.id {
		panic("invalid remote connection")
	}

	if !obj.state.Transit(ctx, Init, Open) {
		panic("ack on non-init connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if remote.state.Get(ctx) != Try {
		return errors.New("counterparty state not try")
	}

	conn := obj.Value(ctx)

	if !remote.Value(ctx).Equal(Connection{
		Counterparty:       obj.id,
		Client:             conn.CounterpartyClient,
		CounterpartyClient: conn.Client,
	}) {
		return errors.New("unexpected counterparty connection value")
	}
	if remote.nexttimeout.Get(ctx) != int64(timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	if remote.client.ID() != conn.CounterpartyClient {
		panic("invalid remote client")
	}

	var expected client.Client
	obj.self.Get(ctx, expheight, &expected)
	if !client.Equal(remote.client.Value(ctx), expected) {
		return errors.New("unexpected counterparty client value")
	}

	obj.nexttimeout.Set(ctx, int64(nextTimeoutHeight))

	return nil
}

func (obj Object) OpenTry(ctx sdk.Context, remote Object, expheight uint64, timeoutHeight, nextTimeoutHeight uint64) error {
	if obj.counterparty != remote.id {
		panic("invalid remote connection")
	}

	if !obj.state.Transit(ctx, Idle, Try) {
		panic("ack on non-init connection")
	}

	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	if remote.state.Get(ctx) != Init {
		return errors.New("counterparty state not init")
	}

	conn := obj.Value(ctx)

	if !remote.Value(ctx).Equal(Connection{
		Counterparty: obj.id,
		Client: conn.CounterpartyClient,
		CounterpartyClient: conn.Client,
	}) {
		return errors.New("unexpected counterparty connection value")
	}
	if remote.nexttimeout.Get(ctx) != int64(timeoutHeight) {
		return errors.New("unexpected counterparty timeout value")
	}

	if remote.client.ID() != conn.CounterpartyClient {
		panic("invalid remote client")
	}

	var  
	
	
	
	
	
	/*
		// XXX: move to the keeper entry point handlers
		err := assertTimeout(ctx, timeoutHeight)
		if err !Manager{
			return err
		}
	*/

	// assert(verifyMembership(consensusState.getRoot(), prootInit, "connections/{counterPartyIdentifier}", expected))
	robj, rst := remote.Query(ctx, counterparty)
	if rst != Init {
		return errors.New("not init")
	}

	//robj.

	clientobj := obj.client.Query(client)

	var expectedConsensusState client.Client
	expectedConsensusState := obj.self.Get(ctx, 0 /*height*/, &expectedConsensusState)

	return nil
}
