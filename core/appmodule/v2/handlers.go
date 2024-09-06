package appmodulev2

import (
	"context"
	"fmt"

	transaction "cosmossdk.io/core/transaction"
)

type (
	// PreMsgHandler is a handler that is executed before Handler. If it errors the execution reverts.
	PreMsgHandler = func(ctx context.Context, msg transaction.Msg) error
	// Handler handles the state transition of the provided message.
	Handler = func(ctx context.Context, msg transaction.Msg) (msgResp transaction.Msg, err error)
	// PostMsgHandler runs after Handler, only if Handler does not error. If PostMsgHandler errors
	// then the execution is reverted.
	PostMsgHandler = func(ctx context.Context, msg, msgResp transaction.Msg) error
)

// PreMsgRouter is a router that allows you to register PreMsgHandlers for specific message types.
type PreMsgRouter interface {
	// RegisterPreHandler will register a specific message handler hooking into the message with
	// the provided name.
	RegisterPreHandler(handler PreMsgHandler)
	// RegisterGlobalPreHandler will register a global message handler hooking into any message
	// being executed.
	RegisterGlobalPreHandler(handler PreMsgHandler)
}

// HasPreMsgHandlers is an interface that modules must implement if they want to register PreMsgHandlers.
type HasPreMsgHandlers interface {
	RegisterPreMsgHandlers(router PreMsgRouter)
}

// RegisterPreHandler is a helper function that modules can use to not lose type safety when registering PreMsgHandler to the
// PreMsgRouter. Example usage:
// ```go
//
//	func (k Keeper) BeforeSend(ctx context.Context, req *types.MsgSend) (*types.QueryBalanceResponse, error) {
//	      ... before send logic ...
//	}
//
//	func (m Module) RegisterPreMsgHandlers(router appmodule.PreMsgRouter) {
//	    appmodule.RegisterPreHandler(router, keeper.BeforeSend)
//	}
//
// ```
func RegisterPreHandler[Req transaction.Msg](
	router PreMsgRouter,
	handler func(ctx context.Context, msg Req) error,
) {
	untypedHandler := func(ctx context.Context, m transaction.Msg) error {
		typed, ok := m.(Req)
		if !ok {
			return fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Req))
		}
		return handler(ctx, typed)
	}
	router.RegisterPreHandler(untypedHandler)
}

// MsgRouter is a router that allows you to register Handlers for specific message types.
type MsgRouter interface {
	RegisterHandler(handler Handler)
}

// HasMsgHandlers is an interface that modules must implement if they want to register Handlers.
type HasMsgHandlers interface {
	RegisterMsgHandlers(router MsgRouter)
}

// RegisterHandler is a helper function that modules can use to not lose type safety when registering handlers to the
// QueryRouter or MsgRouter. Example usage:
// ```go
//
//	func (k Keeper) QueryBalance(ctx context.Context, req *types.QueryBalanceRequest) (*types.QueryBalanceResponse, error) {
//	      ... query logic ...
//	}
//
//	func (m Module) RegisterQueryHandlers(router appmodule.QueryRouter) {
//	    appmodule.RegisterHandler(router, keeper.QueryBalance)
//	}
//
// ```
func RegisterHandler[R MsgRouter, Req, Resp transaction.Msg](
	router R,
	handler func(ctx context.Context, msg Req) (msgResp Resp, err error),
) {
	untypedHandler := func(ctx context.Context, m transaction.Msg) (transaction.Msg, error) {
		typed, ok := m.(Req)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Req))
		}
		return handler(ctx, typed)
	}
	router.RegisterHandler(untypedHandler)
}

// PostMsgRouter is a router that allows you to register PostMsgHandlers for specific message types.
type PostMsgRouter interface {
	// RegisterPostHandler will register a specific message handler hooking after the execution of message with
	// the provided name.
	RegisterPostHandler(handler PostMsgHandler)
	// RegisterGlobalPostHandler will register a global message handler hooking after the execution of any message.
	RegisterGlobalPostHandler(handler PostMsgHandler)
}

// HasPostMsgHandlers is an interface that modules must implement if they want to register PostMsgHandlers.
type HasPostMsgHandlers interface {
	RegisterPostMsgHandlers(router PostMsgRouter)
}

// RegisterPostHandler is a helper function that modules can use to not lose type safety when registering handlers to the
// PostMsgRouter. Example usage:
// ```go
//
//	func (k Keeper) AfterSend(ctx context.Context, req *types.MsgSend, resp *types.MsgSendResponse) error {
//	      ... query logic ...
//	}
//
//	func (m Module) RegisterPostMsgHandlers(router appmodule.PostMsgRouter) {
//	    appmodule.RegisterPostHandler(router, keeper.AfterSend)
//	}
//
// ```
func RegisterPostHandler[Req, Resp transaction.Msg](
	router PostMsgRouter,
	handler func(ctx context.Context, msg Req, msgResp Resp) error,
) {
	untypedHandler := func(ctx context.Context, m, mResp transaction.Msg) error {
		typed, ok := m.(Req)
		if !ok {
			return fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Req))
		}
		typedResp, ok := mResp.(Resp)
		if !ok {
			return fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Resp))
		}
		return handler(ctx, typed, typedResp)
	}
	router.RegisterPostHandler(untypedHandler)
}

// QueryRouter is a router that allows you to register QueryHandlers for specific query types.
type QueryRouter interface {
	Register(handler Handler)
}

// HasQueryHandlers is an interface that modules must implement if they want to register QueryHandlers.
type HasQueryHandlers interface {
	RegisterQueryHandlers(router QueryRouter)
}
