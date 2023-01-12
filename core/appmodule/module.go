package appmodule

import (
	"context"

	"cosmossdk.io/depinject"
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

// HasBeginBlocker is the extension interface that modules should implement to run
// custom logic before transaction processing in a block.
type HasBeginBlocker interface {
	AppModule

	// BeginBlock is a method that will be run before transactions are processed in
	// a block.
	BeginBlock(context.Context) error
}

// HasEndBlocker is the extension interface that modules should implement to run
// custom logic after transaction processing in a block.
type HasEndBlocker interface {
	AppModule

	// EndBlock is a method that will be run after transactions are processed in
	// a block.
	EndBlock(context.Context) error
}
