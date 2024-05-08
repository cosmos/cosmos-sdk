package appmodule

import (
	"context"
	"fmt"
)

type (
	// Handler handles the state transition of the provided message.
	Handler = func(ctx context.Context, msg Message) (msgResp Message, err error)
	// PreMsgHandler is a handler that is executed before Handler.
	// If it errors the execution reverts.
	PreMsgHandler = func(ctx context.Context, msg Message) error
	// PostMsgHandler runs after Handler, only if Handler does not error.
	// If PostMsgHandler errors then the execution is reverted.
	PostMsgHandler = func(ctx context.Context, msg, msgResp Message) error
)

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
func RegisterHandler[Req, Resp Message](
	router interface{ RegisterHandler(string, Handler) },
	handler func(ctx context.Context, msg Req) (msgResp Resp, err error),
) {
	untypedHandler := func(ctx context.Context, m Message) (Message, error) {
		typed, ok := m.(Req)
		if !ok {
			return nil, fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Req))
		}
		return handler(ctx, typed)
	}
	router.RegisterHandler(MessageName[Req](), untypedHandler)
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
func RegisterPreHandler[Req Message](
	router PreMsgRouter,
	handler func(ctx context.Context, msg Req) error,
) {
	untypedHandler := func(ctx context.Context, m Message) error {
		typed, ok := m.(Req)
		if !ok {
			return fmt.Errorf("unexpected type %T, wanted: %T", m, *new(Req))
		}
		return handler(ctx, typed)
	}
	router.Register(MessageName[Req](), untypedHandler)
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
func RegisterPostHandler[Req, Resp Message](
	router PostMsgRouter,
	handler func(ctx context.Context, msg Req, msgResp Resp) error,
) {
	untypedHandler := func(ctx context.Context, m, mResp Message) error {
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
	router.Register(MessageName[Req](), untypedHandler)
}

// msg handler

type PreMsgRouter interface {
	// Register will register a specific message handler hooking into the message with
	// the provided name.
	Register(msgName string, handler PreMsgHandler)
	// RegisterGlobal will register a global message handler hooking into any message
	// being executed.
	RegisterGlobal(handler PreMsgHandler)
}

type HasPreMsgHandlers interface {
	RegisterPreMsgHandlers(router PreMsgRouter)
}

type MsgRouter interface {
	Register(msgName string, handler Handler)
}

// HasMsgHandler is implemented by modules that instead of exposing msg server expose a set of handlers.
type HasMsgHandlers interface {
	// RegisterMsgHandlers is implemented by the module that will register msg handlers.
	RegisterMsgHandlers(router MsgRouter)
}

type PostMsgRouter interface {
	// Register will register a specific message handler hooking after the execution of message with
	// the provided name.
	Register(msgName string, handler PostMsgHandler)
	// RegisterGlobal will register a global message handler hooking after the execution of any message.
	RegisterGlobal(handler PreMsgHandler)
}

// query handler

type QueryRouter interface {
	Register(queryName string, handler Handler)
}

type HasQueryHandlers interface {
	RegisterQueryHandlers(router QueryRouter)
}
