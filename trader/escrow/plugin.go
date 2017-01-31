package escrow

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin/types"
)

// EscrowPlugin is a plugin, storing all state prefixed with it's unique name
type EscrowPlugin struct {
	name   string
	height uint64
}

func New(name string) *EscrowPlugin {
	return &EscrowPlugin{name: name}
}

func (mp EscrowPlugin) Name() string {
	return mp.name
}

// SetOption not supported by EscrowPlugin
func (mp EscrowPlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return fmt.Sprintf("Unknown key: %s", key)
}

// prefix let's us store all our info in a separate name-space
func (mp EscrowPlugin) prefix(store types.KVStore) types.KVStore {
	key := fmt.Sprintf("%s/", mp.name)
	return trader.PrefixStore(store, []byte(key))
}

// parse out which tx we use and then run it
func (mp EscrowPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	tx, err := ParseTx(txBytes)
	if err != nil {
		paybackCtx(ctx).Pay(store)
		return abci.ErrEncodingError
	}

	// the tx only can mess with the escrow data due to the prefix
	res, payback := tx.Apply(mp.prefix(store), ctx, mp.height)
	payback.Pay(store)
	return res
}

// placeholder empty to fulfill interface
func (mp *EscrowPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {}

// track the height for expiration
func (mp *EscrowPlugin) BeginBlock(store types.KVStore, height uint64) {
	mp.height = height
}
func (mp *EscrowPlugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	mp.height = height
	return nil
}
