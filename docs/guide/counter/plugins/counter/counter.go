package counter

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

type CounterPluginState struct {
	Counter   int
	TotalFees types.Coins
}

type CounterTx struct {
	Valid bool
	Fee   types.Coins
}

//--------------------------------------------------------------------------------

type CounterPlugin struct {
}

func (cp *CounterPlugin) Name() string {
	return "counter"
}

func (cp *CounterPlugin) StateKey() []byte {
	return []byte(fmt.Sprintf("CounterPlugin.State"))
}

func New() *CounterPlugin {
	return &CounterPlugin{}
}

func (cp *CounterPlugin) SetOption(store types.KVStore, key, value string) (log string) {
	return ""
}

func (cp *CounterPlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {
	// Decode tx
	var tx CounterTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error()).PrependLog("CounterTx Error: ")
	}

	// Validate tx
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("CounterTx.Valid must be true")
	}
	if !tx.Fee.IsValid() {
		return abci.ErrInternalError.AppendLog("CounterTx.Fee is not sorted or has zero amounts")
	}
	if !tx.Fee.IsNonnegative() {
		return abci.ErrInternalError.AppendLog("CounterTx.Fee must be nonnegative")
	}

	// Did the caller provide enough coins?
	if !ctx.Coins.IsGTE(tx.Fee) {
		return abci.ErrInsufficientFunds.AppendLog("CounterTx.Fee was not provided")
	}

	// TODO If there are any funds left over, return funds.
	// e.g. !ctx.Coins.Minus(tx.Fee).IsZero()
	// ctx.CallerAccount is synced w/ store, so just modify that and store it.

	// Load CounterPluginState
	var cpState CounterPluginState
	cpStateBytes := store.Get(cp.StateKey())
	if len(cpStateBytes) > 0 {
		err = wire.ReadBinaryBytes(cpStateBytes, &cpState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

	// Update CounterPluginState
	cpState.Counter += 1
	cpState.TotalFees = cpState.TotalFees.Plus(tx.Fee)

	// Save CounterPluginState
	store.Set(cp.StateKey(), wire.BinaryBytes(cpState))

	return abci.OK
}

func (cp *CounterPlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (cp *CounterPlugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
}

func (cp *CounterPlugin) EndBlock(store types.KVStore, height uint64) (res abci.ResponseEndBlock) {
	return
}
