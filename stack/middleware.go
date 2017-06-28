package stack

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

// middleware lets us wrap a whole stack up into one Handler
//
// heavily inspired by negroni's design
type middleware struct {
	middleware basecoin.Middleware
	next       basecoin.Handler
}

var _ basecoin.Handler = &middleware{}

func (m *middleware) Name() string {
	return m.middleware.Name()
}

// CheckTx always returns an empty success tx
func (m *middleware) CheckTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (basecoin.Result, error) {
	// make sure we pass in proper context to child
	next := secureCheck(m.next, ctx)
	// set the permissions for this app
	ctx = withApp(ctx, m.Name())
	return m.middleware.CheckTx(ctx, store, tx, next)
}

// DeliverTx always returns an empty success tx
func (m *middleware) DeliverTx(ctx basecoin.Context, store types.KVStore, tx basecoin.Tx) (res basecoin.Result, err error) {
	// make sure we pass in proper context to child
	next := secureDeliver(m.next, ctx)
	// set the permissions for this app
	ctx = withApp(ctx, m.Name())
	return m.middleware.DeliverTx(ctx, store, tx, next)
}

// Stack is the entire application stack
type Stack struct {
	middles          []basecoin.Middleware
	handler          basecoin.Handler
	basecoin.Handler // the compiled version, which we expose
}

var _ basecoin.Handler = &Stack{}

// NewStack prepares a middleware stack, you must `.Use()` a Handler
// before you can execute it.
func NewStack(middlewares ...basecoin.Middleware) *Stack {
	return &Stack{
		middles: middlewares,
	}
}

// Use sets the final handler for the stack and prepares it for use
func (s *Stack) Use(handler basecoin.Handler) *Stack {
	if handler == nil {
		panic("Cannot have a Stack without an end handler")
	}
	s.handler = handler
	s.Handler = build(s.middles, s.handler)
	return s
}

func build(mid []basecoin.Middleware, end basecoin.Handler) basecoin.Handler {
	if len(mid) == 0 {
		return end
	}
	next := build(mid[1:], end)
	return &middleware{mid[0], next}
}
