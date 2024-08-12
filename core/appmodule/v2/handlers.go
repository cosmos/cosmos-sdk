package appmodule

import (
	"context"

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

type HasMsgHandlers interface {
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
