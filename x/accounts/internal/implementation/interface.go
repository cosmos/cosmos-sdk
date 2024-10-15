package implementation

import "context"

// Account defines a smart account interface.
type Account interface {
	// RegisterInitHandler allows the smart account to register an initialisation handler, using
	// the provided InitBuilder. The handler will be called when the smart account is initialized
	// (deployed).
	RegisterInitHandler(builder *InitBuilder)

	// RegisterExecuteHandlers allows the smart account to register execution handlers.
	// The smart account might also decide to not register any execution handler.
	RegisterExecuteHandlers(builder *ExecuteBuilder)

	// RegisterQueryHandlers allows the smart account to register query handlers. The smart account
	// might also decide to not register any query handler.
	RegisterQueryHandlers(builder *QueryBuilder)
}

// AccountExtension is an abstract account extension.
// Currently only used as a marker interface but may become useful in the future to configure extensions while running.
type AccountExtension = interface{}

// ProtoMsgHandlerRegistry abstract registry to register protobuf message handlers of accounts or extensions
type ProtoMsgHandlerRegistry interface {
	RegisterHandler(reqName string, fn ProtoMsgHandler, schema HandlerSchema)
}

var _ ProtoMsgHandlerRegistry = RegisterHandlerFn(nil)

// RegisterHandlerFn adapter to implement ProtoMsgHandlerRegistry
type RegisterHandlerFn func(reqName string, fn ProtoMsgHandler, schema HandlerSchema)

func (r RegisterHandlerFn) RegisterHandler(reqName string, fn ProtoMsgHandler, schema HandlerSchema) {
	r(reqName, fn, schema)
}

// MigrateableLegacyDataExtension is an interface that can be implemented by account extensions to migrate legacy date
type MigrateableLegacyDataExtension interface {
	// MigrateFromLegacy migrate the legacy data from other parts of the system.This method is only called
	// by the accounts module. No external access allowed.
	// The implementation can build upon the account migration that prevents multiple calls.
	// The sender in the context is the module address
	MigrateFromLegacy(ctx context.Context) error
}
