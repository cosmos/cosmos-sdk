package connection

import (
	"errors"

	"github.com/cosmos/cosmos-sdk/store/mapping"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client"
)

type Manager struct {
	m      mapping.Mapping
	client client.Manager
}

func NewManager(protocol, free mapping.Base) Manager {
	return Manager{
		m: mapping.NewMapping(protocol, nil),
	}
}

func (man Manager) object(key string) Object {
	return Object{
		key:        key,
		connection: man.m.Value([]byte(key)),
		state:      man.m.Value([]byte(key + "/state")).Enum(),
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
	client     client.Manager
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

func (obj Object) OpenTry(ctx sdk.Context, counterparty, client, counterpartyClient string /*proofInit Proof*/, timeoutHeight, nextTimeoutHeight uint64) error {
	if ctx.BlockHeight() > int64(timeoutHeight) {
		return errors.New("timeout")
	}

	// TODO
	return nil
}
