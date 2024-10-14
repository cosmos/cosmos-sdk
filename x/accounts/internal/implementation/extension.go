package implementation

import (
	"context"
	"fmt"
	"maps"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"iter"

	"cosmossdk.io/core/transaction"
)

var _ ProtoMsgHandlerRegistry = &ExtensionExecuteAdapter{}

type ProtoMsgHandler func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error)

type ExtensionExecuteAdapter struct {
	// handlers is a map of handler functions that will be called when the smart account is executed.
	handlers map[string]ProtoMsgHandler

	// protoSchemas is a map of schemas for the messages that will be passed to the handler functions
	// and the messages that will be returned by the handler functions.
	protoSchemas map[string]HandlerSchema
}

func NewExtensionExecuteAdapter() *ExtensionExecuteAdapter {
	return &ExtensionExecuteAdapter{
		handlers:     make(map[string]ProtoMsgHandler),
		protoSchemas: make(map[string]HandlerSchema),
	}
}

func (e ExtensionExecuteAdapter) ExtendHandlers(accountExec ProtoMsgHandler) (ProtoMsgHandler, error) {
	return func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error) {
		if len(e.handlers) != 0 {
			messageName := MessageName(msg)
			handler, ok := e.handlers[messageName]
			if ok {
				return handler(ctx, msg)
			}
		}
		return accountExec(ctx, msg)
	}, nil
}

func (e *ExtensionExecuteAdapter) RegisterExtension(deps Dependencies, setup AccountExtensionCreatorFunc) error {
	_, _, err := setup(deps, e)
	return err
}

// RegisterHandler implements ProtoMsgHandlerRegistry
func (e *ExtensionExecuteAdapter) RegisterHandler(reqName string, fn ProtoMsgHandler, schema HandlerSchema) {
	// check if not registered already
	if _, ok := e.handlers[reqName]; ok {
		panic(fmt.Sprintf("handler already registered for message %s", reqName))
	}
	e.handlers[reqName] = fn
	e.protoSchemas[reqName] = schema
}

type InterfaceRegistry interface {
	RegisterInterface(name string, iface any, impls ...gogoproto.Message)
	RegisterImplementations(iface any, impls ...gogoproto.Message)
}

func (e ExtensionExecuteAdapter) HandlerSchemas() iter.Seq[HandlerSchema] {
	return maps.Values(e.protoSchemas)
}

// ProtoMsgHandlerRegistry abstract registry to register protobuf message handlers of accounts or extensions
type ProtoMsgHandlerRegistry interface {
	RegisterHandler(reqName string, fn ProtoMsgHandler, schema HandlerSchema)
}
