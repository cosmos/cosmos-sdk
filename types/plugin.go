package types

import (
	abci "github.com/tendermint/abci/types"
)

type Plugin interface {
	SetOption(store KVStore, key string, value string) (log string)
	RunTx(store KVStore, ctx CallContext, txBytes []byte) (res abci.Result)
	InitChain(store KVStore, vals []*abci.Validator)
	BeginBlock(store KVStore, height uint64)
	EndBlock(store KVStore, height uint64) []*abci.Validator
}

type NamedPlugin struct {
	Name string
	Plugin
}

//----------------------------------------

type CallContext struct {
	Caller []byte
	Coins  Coins
}

func NewCallContext(caller []byte, coins Coins) CallContext {
	return CallContext{
		Caller: caller,
		Coins:  coins,
	}
}

//----------------------------------------

type Plugins struct {
	byName map[string]Plugin
	plist  []NamedPlugin
}

func NewPlugins() *Plugins {
	return &Plugins{
		byName: make(map[string]Plugin),
	}
}

func (pgz *Plugins) RegisterPlugin(name string, plugin Plugin) {
	pgz.byName[name] = plugin
	pgz.plist = append(pgz.plist, NamedPlugin{
		Name:   name,
		Plugin: plugin,
	})
}

func (pgz *Plugins) GetByName(name string) Plugin {
	return pgz.byName[name]
}

func (pgz *Plugins) GetList() []NamedPlugin {
	return pgz.plist
}
