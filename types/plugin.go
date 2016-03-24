package types

import (
	tmsp "github.com/tendermint/tmsp/types"
)

// Value is any floating value.  It must be given to someone.
// Gas is a pointer to remainig gas.  Decrement as you go,
// if any gas is left the user is
type Plugin interface {
	CallTx(ctx CallContext, txBytes []byte) tmsp.Result
}

type CallContext struct {
	Cache  AccountCacher
	Caller *Account
	Value  int64
	Gas    *int64
}

func NewCallContext(cache AccountCacher, caller *Account, value int64, gas *int64) CallContext {
	return CallContext{
		Cache:  cache,
		Caller: caller,
		Value:  value,
		Gas:    gas,
	}
}
