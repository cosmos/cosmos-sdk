package types

import (
	tmsp "github.com/tendermint/tmsp/types"
)

type Plugin interface {
	SetOption(store KVStore, key string, value string) (log string)
	RunTx(store KVStore, ctx CallContext, txBytes []byte) (res tmsp.Result)
	InitChain(store KVStore, vals []*tmsp.Validator)
	BeginBlock(store KVStore, height uint64)
	EndBlock(store KVStore, height uint64) []*tmsp.Validator
}

type NamedPlugin struct {
	Byte byte
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
	byByte map[byte]Plugin
	byName map[string]Plugin
	plist  []NamedPlugin
}

func NewPlugins() *Plugins {
	return &Plugins{
		byByte: make(map[byte]Plugin),
		byName: make(map[string]Plugin),
	}
}

func (pgz *Plugins) RegisterPlugin(typeByte byte, name string, plugin Plugin) {
	pgz.byByte[typeByte] = plugin
	pgz.byName[name] = plugin
	pgz.plist = append(pgz.plist, NamedPlugin{
		Byte:   typeByte,
		Name:   name,
		Plugin: plugin,
	})
}

func (pgz *Plugins) GetByByte(typeByte byte) Plugin {
	return pgz.byByte[typeByte]
}

func (pgz *Plugins) GetByName(name string) Plugin {
	return pgz.byName[name]
}

func (pgz *Plugins) GetList() []NamedPlugin {
	return pgz.plist
}
