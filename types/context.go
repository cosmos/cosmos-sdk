package types

import (
	"context"
	abci "github.com/tendermint/abci/types"
)

type Context struct {
	context.Context
	// Don't add any other fields here,
	// it's probably not what you want to do.
}

func NewContext(header tm.Header, isCheckTx bool, txBytes []byte) Context {
	c := Context{
		Context: context.Background(),
	}
	c = c.setBlockHeader(header)
	c = c.setBlockHeight(int64(header.Height))
	c = c.setChainID(header.ChainID)
	c = c.setIsCheckTx(isCheckTx)
	c = c.setTxBytes(txBytes)
	return c
}

// The original context.Context API.
func (c Context) WithValue(key interface{}, value interface{}) context.Context {
	return context.WithValue(c.Context, key, value)
}

// Like WithValue() but retains this API.
func (c Context) WithValueSDK(key interface{}, value interface{}) Context {
	return Context{
		Context: context.WithValue(c.Context, key, value),
	}
}

//----------------------------------------
// Our extensions

type contextKey int // local to the context module

const (
	contextKeyBlockHeader contextKey = iota
	contextKeyBlockHeight
	contextKeyChainID
	contextKeyIsCheckTx
	contextKeyTxBytes
)

func (c Context) BlockHeader() tm.Header {
	return c.Value(contextKeyBlockHeader).(tm.Header)
}

func (c Context) BlockHeight() int64 {
	return c.Value(contextKeyBlockHeight).(int64)
}

func (c Context) ChainID() string {
	return c.Value(contextKeyChainID).(string)
}

func (c Context) IsCheckTx() bool {
	return c.Value(contextKeyIsCheckTx).(bool)
}

func (c Context) TxBytes() []byte {
	return c.Value(contextKeyTxBytes).([]byte)
}

// Unexposed to prevent overriding.
func (c Context) setBlockHeader(header tm.Header) Context {
	return c.WithValueSDK(contextKeyBlockHeader, header)
}

// Unexposed to prevent overriding.
func (c Context) setBlockHeight(height int64) Context {
	return c.WithValueSDK(contextKeyBlockHeight, header)
}

// Unexposed to prevent overriding.
func (c Context) setChainID(chainID string) Context {
	return c.WithValueSDK(contextKeyChainID, header)
}

// Unexposed to prevent overriding.
func (c Context) setIsCheckTx(isCheckTx bool) Context {
	return c.WithValueSDK(contextKeyIsCheckTx, isCheckTx)
}

// Unexposed to prevent overriding.
func (c Context) setTxBytes(txBytes []byte) Context {
	return c.WithValueSDK(contextKeyTxBytes, txBytes)
}
