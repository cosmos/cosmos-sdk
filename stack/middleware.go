package stack

import (
	"github.com/tendermint/tmlibs/log"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/state"
)

// middleware lets us wrap a whole stack up into one Handler
//
// heavily inspired by negroni's design
type middleware struct {
	middleware Middleware
	space      string
	allowIBC   bool
	next       basecoin.Handler
}

var _ basecoin.Handler = &middleware{}

func (m *middleware) Name() string {
	return m.middleware.Name()
}

func (m *middleware) wrapCtx(ctx basecoin.Context) basecoin.Context {
	if m.allowIBC {
		return withIBC(ctx)
	}
	return withApp(ctx, m.space)
}

// CheckTx always returns an empty success tx
func (m *middleware) CheckTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (basecoin.Result, error) {
	// make sure we pass in proper context to child
	next := secureCheck(m.next, ctx)
	// set the permissions for this app
	ctx = m.wrapCtx(ctx)
	store = stateSpace(store, m.space)

	return m.middleware.CheckTx(ctx, store, tx, next)
}

// DeliverTx always returns an empty success tx
func (m *middleware) DeliverTx(ctx basecoin.Context, store state.SimpleDB, tx basecoin.Tx) (res basecoin.Result, err error) {
	// make sure we pass in proper context to child
	next := secureDeliver(m.next, ctx)
	// set the permissions for this app
	ctx = m.wrapCtx(ctx)
	store = stateSpace(store, m.space)

	return m.middleware.DeliverTx(ctx, store, tx, next)
}

func (m *middleware) SetOption(l log.Logger, store state.SimpleDB, module, key, value string) (string, error) {
	// set the namespace for the app
	store = stateSpace(store, m.space)

	return m.middleware.SetOption(l, store, module, key, value, m.next)
}

// builder is used to associate info with the middleware, so we can build
// it properly
type builder struct {
	middleware Middleware
	stateSpace string
	allowIBC   bool
}

func prep(m Middleware, ibc bool) builder {
	return builder{
		middleware: m,
		stateSpace: m.Name(),
		allowIBC:   ibc,
	}
}

// wrap sets up the middleware with the proper options
func (b builder) wrap(next basecoin.Handler) basecoin.Handler {
	return &middleware{
		middleware: b.middleware,
		space:      b.stateSpace,
		allowIBC:   b.allowIBC,
		next:       next,
	}
}

// Stack is the entire application stack
type Stack struct {
	middles          []builder
	handler          basecoin.Handler
	basecoin.Handler // the compiled version, which we expose
}

var _ basecoin.Handler = &Stack{}

// New prepares a middleware stack, you must `.Use()` a Handler
// before you can execute it.
func New(middlewares ...Middleware) *Stack {
	stack := new(Stack)
	return stack.Apps(middlewares...)
}

// Apps adds the following Middlewares as typical application
// middleware to the stack (limit permission to one app)
func (s *Stack) Apps(middlewares ...Middleware) *Stack {
	// TODO: some wrapper...
	for _, m := range middlewares {
		s.middles = append(s.middles, prep(m, false))
	}
	return s
}

// IBC add the following middleware with permission to add cross-chain
// permissions
func (s *Stack) IBC(m Middleware) *Stack {
	// TODO: some wrapper...
	s.middles = append(s.middles, prep(m, true))
	return s
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

// Dispatch is like Use, but a convenience method to construct a
// dispatcher with a set of modules to route.
func (s *Stack) Dispatch(routes ...Dispatchable) *Stack {
	d := NewDispatcher(routes...)
	return s.Use(d)
}

func build(mid []builder, end basecoin.Handler) basecoin.Handler {
	if len(mid) == 0 {
		return end
	}
	next := build(mid[1:], end)
	return mid[0].wrap(next)
}
