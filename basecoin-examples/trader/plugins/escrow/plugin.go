package escrow

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin-examples/trader/types"
	bc "github.com/tendermint/basecoin/types"
)

// Plugin is a plugin, storing all state prefixed with it's unique name
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
func (p Plugin) SetOption(store bc.KVStore, key string, value string) (log string) {
	return fmt.Sprintf("Unknown key: %s", key)
}

// prefix let's us store all our info in a separate name-space
func (p Plugin) prefix(store bc.KVStore) bc.KVStore {
	key := fmt.Sprintf("%s/", p.name)
	return trader.PrefixStore(store, []byte(key))
}

// parse out which tx we use and then run it
func (p Plugin) RunTx(store bc.KVStore, ctx bc.CallContext, txBytes []byte) (res abci.Result) {
	tx, err := types.ParseEscrowTx(txBytes)
	if err != nil {
		trader.NewAccountant(store).Refund(ctx)
		return abci.ErrEncodingError
	}

	// the tx only can mess with the escrow data due to the prefix
	res = p.Exec(store, ctx, tx)
	return res
}

func (p Plugin) Exec(store bc.KVStore, ctx bc.CallContext, tx types.EscrowTx) abci.Result {
	accts := trader.NewAccountant(store)
	pstore := p.prefix(store)

	switch t := tx.(type) {
	case types.CreateEscrowTx:
		return p.runCreateEscrow(pstore, accts, ctx, t)
	case types.ResolveEscrowTx:
		return p.runResolveEscrow(pstore, accts, ctx, t)
	case types.ExpireEscrowTx:
		return p.runExpireEscrow(pstore, accts, ctx, t)
	default:
		return abci.ErrUnknownRequest
	}
}

// placeholder empty to fulfill interface
func (p *Plugin) InitChain(store bc.KVStore, vals []*abci.Validator) {}

// track the height for expiration
func (p *Plugin) BeginBlock(store bc.KVStore, hash []byte, header *abci.Header) {
	p.height = header.Height
}

func (p *Plugin) EndBlock(store bc.KVStore, height uint64) abci.ResponseEndBlock {
	p.height = height + 1
	return abci.ResponseEndBlock{}
}

func (p *Plugin) assertPlugin() bc.Plugin {
	return p
}
