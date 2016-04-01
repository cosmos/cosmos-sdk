package types

import (
	tmsp "github.com/tendermint/tmsp/types"
)

// Value is any floating value.  It must be given to someone.
type Plugin interface {
	SetOption(key string, value string) (log string)
	RunTx(ctx CallContext, txBytes []byte) (res tmsp.Result)
	Query(query []byte) (res tmsp.Result)
	Commit() (res tmsp.Result)
}

type NamedPlugin struct {
	Byte byte
	Name string
	Plugin
}

//----------------------------------------

type CallContext struct {
	Cache  AccountCacher
	Caller *Account
	Coins  Coins
}

func NewCallContext(cache AccountCacher, caller *Account, coins Coins) CallContext {
	return CallContext{
		Cache:  cache,
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
