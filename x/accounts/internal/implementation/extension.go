package implementation

import (
	"context"
	"fmt"
	"iter"
	"maps"

	"cosmossdk.io/core/transaction"
)

type ProtoMsgHandler func(ctx context.Context, msg transaction.Msg) (resp transaction.Msg, err error)

type ExtensionHandlers struct {
	// handlers is a map of handler functions that will be called when the smart account is executed.
	handlers map[string]ProtoMsgHandler

	// protoSchemas is a map of schemas for the messages that will be passed to the handler functions
	// and the messages that will be returned by the handler functions.
	protoSchemas map[string]HandlerSchema
	// legacy data migrations
	dataMigrations []DataMigrationExtension
}

func NewExtensionHandlers() *ExtensionHandlers {
	return &ExtensionHandlers{
		handlers:       make(map[string]ProtoMsgHandler),
		protoSchemas:   make(map[string]HandlerSchema),
		dataMigrations: make([]DataMigrationExtension, 0),
	}
}

func (e ExtensionHandlers) ExtendHandlers(accountExec ProtoMsgHandler) (ProtoMsgHandler, error) {
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

// RegisterExtension instantiate extension from construtor function
func (e *ExtensionHandlers) RegisterExtension(deps Dependencies, constructor AccountExtensionCreatorFunc) error {
	reg := RegisterHandlerFn(func(reqName string, fn ProtoMsgHandler, schema HandlerSchema) {
		// check if not registered already
		if _, ok := e.handlers[reqName]; ok {
			panic(fmt.Sprintf("handler already registered for message %s", reqName))
		}
		e.handlers[reqName] = fn
		e.protoSchemas[reqName] = schema
	})
	name, obj, err := constructor(deps, reg)
	x, ok := obj.(MigrateableLegacyDataExtension)
	if ok {
		e.dataMigrations = append(e.dataMigrations, DataMigrationExtension{
			name: name,
			exec: x,
		})
	}
	return err
}

// HandlerSchemas returns iterator to access all schemas. Not thread safe
func (e ExtensionHandlers) HandlerSchemas() iter.Seq[HandlerSchema] {
	return maps.Values(e.protoSchemas)
}

func (e ExtensionHandlers) MigrateLegacyState(ctx context.Context) error {
	for _, m := range e.dataMigrations {
		if err := m.exec.MigrateFromLegacy(ctx); err != nil {
			return fmt.Errorf("migrate legacy state failed for extension %q: %w", m.name, err)
		}
	}
	return nil
}

type DataMigrationExtension struct {
	name string
	exec MigrateableLegacyDataExtension
}
