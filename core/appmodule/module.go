package appmodule

import (
	"context"

	"cosmossdk.io/depinject"
	"google.golang.org/grpc"
)

// AppModule is a tag interface for app module implementations to use as a basis
// for extension interfaces. It provides no functionality itself, but is the
// type that all valid app modules should provide so that they can be identified
// by other modules (usually via depinject) as app modules.
type AppModule interface {
	depinject.OnePerModuleType

	// IsAppModule is a dummy method to tag a struct as implementing an AppModule.
	IsAppModule()
}

type HasServices interface {
	AppModule

	RegisterServices(grpc.ServiceRegistrar)
}

type HasBeginBlocker interface {
	BeginBlock(context.Context) error
}

type HasEndBlocker interface {
	EndBlock(context.Context) error
}

type HasGenesis interface {
	DefaultGenesis(GenesisTarget)
	ValidateGenesis(GenesisSource) error
	InitGenesis(context.Context, GenesisSource) error
	ExportGenesis(context.Context, GenesisTarget)
}
