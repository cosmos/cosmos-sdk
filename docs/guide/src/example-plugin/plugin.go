package main

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

type ExamplePluginState struct {
	Counter int
}

type ExamplePluginTx struct {
	Valid bool
}

type ExamplePlugin struct {
	name string
}

func (ep *ExamplePlugin) Name() string {
	return ep.name
}

func (ep *ExamplePlugin) StateKey() []byte {
	return []byte("ExamplePlugin.State")
}

func NewExamplePlugin() *ExamplePlugin {
	return &ExamplePlugin{
		name: "example-plugin",
	}
}

func (ep *ExamplePlugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (ep *ExamplePlugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var tx ExamplePluginTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	if !tx.Valid {
		return abci.ErrInternalError.AppendLog("Valid must be true")
	}

	// Load PluginState
	var pluginState ExamplePluginState
	stateBytes := store.Get(ep.StateKey())
	if len(stateBytes) > 0 {
		err = wire.ReadBinaryBytes(stateBytes, &pluginState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

	//App Logic
	pluginState.Counter += 1

	// Save PluginState
	store.Set(ep.StateKey(), wire.BinaryBytes(pluginState))

	return abci.OK
}

func (cp *ExamplePlugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (ep *ExamplePlugin) BeginBlock(store types.KVStore, hash []byte, header *abci.Header) {
}

func (ep *ExamplePlugin) EndBlock(store types.KVStore, height uint64) abci.ResponseEndBlock {
	return abci.ResponseEndBlock{}
}
