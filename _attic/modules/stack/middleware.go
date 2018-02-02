package stack

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

// middleware lets us wrap a whole stack up into one Handler
//
// heavily inspired by negroni's design
type middleware struct {
	middleware Middleware
	space      string
	allowIBC   bool
	next       sdk.Handler
}

var _ sdk.Handler = &middleware{}

func (m *middleware) Name() string {
	return m.middleware.Name()
}

func (m *middleware) wrapCtx(ctx sdk.Context) sdk.Context {
	if m.allowIBC {
		return withIBC(ctx)
	}
	return withApp(ctx, m.space)
}

// CheckTx always returns an empty success tx
func (m *middleware) CheckTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (sdk.CheckResult, error) {
	// make sure we pass in proper context to child
	next := secureCheck(m.next, ctx)
	// set the permissions for this app
	ctx = m.wrapCtx(ctx)
	store = stateSpace(store, m.space)

	return m.middleware.CheckTx(ctx, store, tx, next)
}

// DeliverTx always returns an empty success tx
func (m *middleware) DeliverTx(ctx sdk.Context, store state.SimpleDB, tx sdk.Tx) (res sdk.DeliverResult, err error) {
	// make sure we pass in proper context to child
	next := secureDeliver(m.next, ctx)
	// set the permissions for this app
	ctx = m.wrapCtx(ctx)
	store = stateSpace(store, m.space)

	return m.middleware.DeliverTx(ctx, store, tx, next)
}

func (m *middleware) InitState(l log.Logger, store state.SimpleDB, module, key, value string) (string, error) {
	// set the namespace for the app
	store = stateSpace(store, m.space)

	return m.middleware.InitState(l, store, module, key, value, m.next)
}

func (m *middleware) InitValidate(l log.Logger, store state.SimpleDB, vals []*abci.Validator) {
	// set the namespace for the app
	store = stateSpace(store, m.space)
	m.middleware.InitValidate(l, store, vals, m.next)
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
func (b builder) wrap(next sdk.Handler) sdk.Handler {
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
	handler          sdk.Handler
	sdk.Handler // the compiled version, which we expose
}

var _ sdk.Handler = &Stack{}

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
func (s *Stack) Use(handler sdk.Handler) *Stack {
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

func build(mid []builder, end sdk.Handler) sdk.Handler {
	if len(mid) == 0 {
		return end
	}
	next := build(mid[1:], end)
	return mid[0].wrap(next)
}
