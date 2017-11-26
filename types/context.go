package types

import (
	"context"
	tm "github.com/tendermint/tendermint/types"
)

/*

NOTE: Golang's Context is embedded and relied on
for compatibility w/ tools like monkit.
(https://github.com/spacemonkeygo/monkit)

Usage:

defer mon.Task()(&ctx.Context)(&err)

*/
type SDKContext struct {
	context.Context
	// NOTE: adding fields here will break monkit compatibility
	// use context.Context instead if possible.
}

func NewSDKContext(header tm.Header) SDKContext {
	c := SDKContext{
		Context: context.Background(),
	}
	c = c.setBlockHeader(header)
	c = c.setBlockHeight(int64(header.Height))
	c = c.setChainID(header.ChainID)
	return c
}

func (c SDKContext) WithValueSDK(key interface{}, value interface{}) SDKContext {
	return SDKContext{
		Context: context.WithValue(c.Context, key, value),
	}
}

func (c SDKContext) WithValue(key interface{}, value interface{}) Context {
	return c
}

//----------------------------------------
// Our extensions

type contextKey int // local to the context module

const (
	contextKeyBlockHeader contextKey = iota
	contextKeyBlockHeight
	contextKeyChainID
)

func (c SDKContext) BlockHeader() tm.Header {
	return c.Value(contextKeyBlockHeader).(tm.Header)
}

// Unexposed to prevent overriding.
func (c SDKContext) setBlockHeader(header tm.Header) SDKContext {
	return c.WithValueSDK(contextKeyBlockHeader, header)
}

// Unexposed to prevent overriding.
func (c SDKContext) setBlockHeight(height int64) SDKContext {
	return c.WithValueSDK(contextKeyBlockHeight, header)
}

// Unexposed to prevent overriding.
func (c SDKContext) setChainID(chainID string) SDKContext {
	return c.WithValueSDK(contextKeyChainID, header)
}
