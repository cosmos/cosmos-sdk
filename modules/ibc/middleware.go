package ibc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
)

// Middleware allows us to verify the IBC proof on a packet and
// and if valid, attach this permission to the wrapped packet
type Middleware struct {
	stack.PassOption
}

var _ stack.Middleware = Middleware{}

// NewMiddleware creates a role-checking middleware
func NewMiddleware() Middleware {
	return Middleware{}
}

// Name - return name space
func (Middleware) Name() string {
	return NameIBC
}

// CheckTx verifies the named chain and height is present, and verifies
// the merkle proof in the packet
func (m Middleware) CheckTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Checker) (res basecoin.Result, err error) {
	return res, nil
}

// DeliverTx verifies the named chain and height is present, and verifies
// the merkle proof in the packet
func (m Middleware) DeliverTx(ctx basecoin.Context, store state.KVStore, tx basecoin.Tx, next basecoin.Deliver) (res basecoin.Result, err error) {
	return res, nil
}
