package options

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
)

// Plugin is a options plugin, storing all state prefixed with it's unique name
type Plugin struct {
	name   string
	height uint64
}

func New(name string) *Plugin {
	return &Plugin{name: name}
}

func (p Plugin) Name() string {
	return p.name
}

// SetOption not supported by Plugin
func (p Plugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return fmt.Sprintf("Unknown key: %s", key)
}

// prefix let's us store all our info in a separate name-space
func (p Plugin) prefix(store types.KVStore) types.KVStore {
	key := fmt.Sprintf("%s/", p.name)
	return trader.PrefixStore(store, []byte(key))
}

// parse out which tx we use and then run it
func (p Plugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	tx, err := ParseTx(txBytes)
	if err != nil {
		// TODO: how to pay back???
		// paybackCtx(ctx).Pay(store)
		return abci.ErrEncodingError
	}

	// the tx only can mess with the escrow data due to the prefix
	tstore := p.prefix(store)
	accts := state.NewState(store)
	res = tx.Apply(tstore, accts, ctx, p.height)
	// TODO: how to pay back???
	// payback.Pay(store)
	return res
}

// placeholder empty to fulfill interface
func (p *Plugin) InitChain(store types.KVStore, vals []*abci.Validator) {}

// track the height for expiration
func (p *Plugin) BeginBlock(store types.KVStore, height uint64) {
	p.height = height
}
func (p *Plugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	p.height = height
	return nil
}
