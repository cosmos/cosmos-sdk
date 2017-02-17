package types

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
)

type Plugin interface {

	// Name of this plugin, should be short.
	Name() string

	// Run a transaction from ABCI DeliverTx
	RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)

	// Other ABCI message handlers
	SetOption(store KVStore, key, value string) (log string)
	InitChain(store KVStore, vals []*abci.Validator)
	BeginBlock(store KVStore, hash []byte, header *abci.Header)
	EndBlock(store KVStore, height uint64) abci.ResponseEndBlock
}

//----------------------------------------

type CallContext struct {
	CallerAddress []byte   // Caller's Address (hash of PubKey)
	CallerAccount *Account // Caller's Account, w/ fee & TxInputs deducted
	Coins         Coins    // The coins that the caller wishes to spend, excluding fees
}

func NewCallContext(callerAddress []byte, callerAccount *Account, coins Coins) CallContext {
	return CallContext{
		CallerAddress: callerAddress,
		CallerAccount: callerAccount,
		Coins:         coins,
	}
}

//----------------------------------------

type Plugins struct {
	byName map[string]Plugin
	plist  []Plugin
}

func NewPlugins() *Plugins {
	return &Plugins{
		byName: make(map[string]Plugin),
	}
}

func (pgz *Plugins) RegisterPlugin(plugin Plugin) {
	name := plugin.Name()
	if name == "" {
		panic("Plugin name cannot be blank")
	}
	if _, exists := pgz.byName[name]; exists {
		panic(fmt.Sprintf("Plugin already exists by the name of %v", name))
	}
	pgz.byName[name] = plugin
	pgz.plist = append(pgz.plist, plugin)
}

func (pgz *Plugins) GetByName(name string) Plugin {
	return pgz.byName[name]
}

func (pgz *Plugins) GetList() []Plugin {
	return pgz.plist
}
