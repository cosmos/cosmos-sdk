package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

type Manager struct {
	protocol mapping.Mapping

	states mapping.Indexer

	client client.Manager
}

func NewManager(protocol, free mapping.Base, client client.Manager) Manager {
	return Manager{
		protocol: mapping.NewMapping(protocol, []byte("/")),

		states: mapping.NewIndexer(free, []byte("/states"), mapping.Dec),

		client: client,
	}
}

func (man Manager) object(key string) Object {
	return Object{
		key:        key,
		connection: man.protocol.Value([]byte(key)),
		state:      man.protocol.Value([]byte(key + "/state")).Enum(),

		states: man.states,
		client: man.client,
	}
}

func (man Manager) Query(ctx sdk.Context, key string) (res Object, st State) {
	res = man.object(key)
	st = res.state.Get(ctx)
	return
}

type Object struct {
	key        string
	connection mapping.Value
	state      mapping.Enum

	states mapping.Indexer
	client client.Manager
}

func (obj Object) exists(ctx sdk.Context) bool {
	return obj.connection.Exists(ctx)
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
		NextTimeoutHeight:  nextTimeoutHeight,
	})

	return nil
}

func assertTimeout(ctx sdk.Context, timeoutHeight uint) error {
	if ctx.BlockHeight() > int64(timeoutHeight) {
		return errors.New("timeout")
	}

	return nil
}

func (obj Object) OpenTry(ctx sdk.Context, counterparty, client, counterpartyClient string /*proofInit Proof*/, timeoutHeight, nextTimeoutHeight uint64) error {
	err := assertTimeout(ctx, timeoutHeight)
	if err != nil {
		return err
	}

	clientobj := obj.client.Query(client)

	// TODO

	return nil
}
