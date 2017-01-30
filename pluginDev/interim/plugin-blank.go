package Blank

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
)

type PluginState struct {
}

type PluginTx struct {
}

type Plugin struct {
	name string
}

func (cp *Plugin) Name() string {
	return Plugin.name
}

func (cp *Plugin) StateKey() []byte {
	return []byte(fmt.Sprintf("Plugin{name=%v}.State", cp.name))
}

func New(name string) *Plugin {
	return &Plugin{
		name: name,
	}
}

func (cp *Plugin) SetOption(store types.KVStore, key string, value string) (log string) {
	return ""
}

func (cp *Plugin) RunTx(store types.KVStore, ctx types.CallContext, txBytes []byte) (res abci.Result) {

	// Decode tx
	var tx CounterTx
	err := wire.ReadBinaryBytes(txBytes, &tx)
	if err != nil {
		return abci.ErrBaseEncodingError.AppendLog("Error decoding tx: " + err.Error())
	}

	// Validate tx
	//TODO: validate transaction logic

	// Load PluginState
	var cpState PluginState
	cpStateBytes := store.Get(cp.StateKey())
	if len(cpStateBytes) > 0 {
		err = wire.ReadBinaryBytes(cpStateBytes, &cpState)
		if err != nil {
			return abci.ErrInternalError.AppendLog("Error decoding state: " + err.Error())
		}
	}

	//App Logic
	//TODO: Perform Application logic and update plugin state

	// Save PluginState
	store.Set(cp.StateKey(), wire.BinaryBytes(cpState))

	return abci.OK
}

func (cp *Plugin) InitChain(store types.KVStore, vals []*abci.Validator) {
}

func (cp *Plugin) BeginBlock(store types.KVStore, height uint64) {
}

func (cp *Plugin) EndBlock(store types.KVStore, height uint64) []*abci.Validator {
	return nil
}
