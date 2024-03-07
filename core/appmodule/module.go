package appmodule

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/runtime/protoiface"

	"cosmossdk.io/core/appmodule/v2"
)

// AppModule is a tag interface for app module implementations to use as a basis
// for extension interfaces. It provides no functionality itself, but is the
// type that all valid app modules should provide so that they can be identified
// by other modules (usually via depinject) as app modules.
type AppModule = appmodule.AppModule

// HasMigrations is the extension interface that modules should implement to register migrations.
type HasMigrations interface {
	AppModule

	// RegisterMigrations registers the module's migrations with the app's migrator.
	RegisterMigrations(MigrationRegistrar) error
}

// HasConsensusVersion is the interface for declaring a module consensus version.
type HasConsensusVersion = appmodule.HasConsensusVersion

// HasServices is the extension interface that modules should implement to register
// implementations of services defined in .proto files.
type HasServices interface {
	AppModule

	// RegisterServices registers the module's services with the app's service
	// registrar.
	//
	// Two types of services are currently supported:
	// - read-only gRPC query services, which are the default.
	// - transaction message services, which must have the protobuf service
	//   option "cosmos.msg.v1.service" (defined in "cosmos/msg/v1/service.proto")
	//   set to true.
	//
	// The service registrar will figure out which type of service you are
	// implementing based on the presence (or absence) of protobuf options. You
	// do not need to specify this in golang code.
	RegisterServices(grpc.ServiceRegistrar) error
}

// ResponsePreBlock represents the response from the PreBlock method.
// It can modify consensus parameters in storage and signal the caller through the return value.
// When it returns ConsensusParamsChanged=true, the caller must refresh the consensus parameter in the finalize context.
// The new context (ctx) must be passed to all the other lifecycle methods.
type ResponsePreBlock interface {
	IsConsensusParamsChanged() bool
}

// HasPreBlocker is the extension interface that modules should implement to run
// custom logic before BeginBlock.
type HasPreBlocker interface {
	AppModule
	// PreBlock is method that will be run before BeginBlock.
	PreBlock(context.Context) (ResponsePreBlock, error)
}

// HasBeginBlocker is the extension interface that modules should implement to run
// custom logic before transaction processing in a block.
type HasBeginBlocker = appmodule.HasBeginBlocker

// HasEndBlocker is the extension interface that modules should implement to run
// custom logic after transaction processing in a block.
type HasEndBlocker = appmodule.HasEndBlocker

// HasRegisterInterfaces is the interface for modules to register their msg types.
type HasRegisterInterfaces = appmodule.HasRegisterInterfaces

// MsgHandlerRouter is implemented by the runtime provider.
type MsgHandlerRouter interface {
	// RegisterHandler is called by modules to register msg handler functions.
	RegisterHandler(name string, handler func(ctx context.Context, msg protoiface.MessageV1) (msgResp protoiface.MessageV1, err error))
}

// HasMsgHandler is implemented by modules that instead of exposing msg server expose
// a set of handlers.
type HasMsgHandler interface {
	// RegisterMsgHandlers is implemented by the module that will register msg handlers.
	RegisterMsgHandlers(router MsgHandlerRouter)
}

// ---------------------------------------------------------------------------- //

// HasPrepareCheckState is an extension interface that contains information about the AppModule
// and PrepareCheckState.
type HasPrepareCheckState interface {
	AppModule
	PrepareCheckState(context.Context) error
}

// HasPrecommit is an extension interface that contains information about the AppModule and Precommit.
type HasPrecommit interface {
	AppModule
	Precommit(context.Context) error
}
