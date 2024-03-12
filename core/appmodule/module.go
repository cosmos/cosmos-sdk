package appmodule

import (
	"context"

	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule/v2"
)

// AppModule is a tag interface for app module implementations to use as a basis
// for extension interfaces. It provides no functionality itself, but is the
// type that all valid app modules should provide so that they can be identified
// by other modules (usually via depinject) as app modules.
type AppModule = appmodule.AppModule

// HasPreBlocker is the extension interface that modules should implement to run
// custom logic before BeginBlock.
type HasPreBlocker = appmodule.HasPreBlocker

// HasBeginBlocker is the extension interface that modules should implement to run
// custom logic before transaction processing in a block.
type HasBeginBlocker = appmodule.HasBeginBlocker

// HasEndBlocker is the extension interface that modules should implement to run
// custom logic after transaction processing in a block.
type HasEndBlocker = appmodule.HasEndBlocker

// HasRegisterInterfaces is the interface for modules to register their msg types.
type HasRegisterInterfaces = appmodule.HasRegisterInterfaces

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

// HasPrepareCheckState is an extension interface that contains information about the AppModule
// and PrepareCheckState.
type HasPrepareCheckState interface {
	appmodule.AppModule
	PrepareCheckState(context.Context) error
}

// HasPrecommit is an extension interface that contains information about the appmodule.AppModule and Precommit.
type HasPrecommit interface {
	appmodule.AppModule
	Precommit(context.Context) error
}
