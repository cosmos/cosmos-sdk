package appmodule

import (
	"context"

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

// ValidatorUpdate defines a validator update.
type ValidatorUpdate = appmodule.ValidatorUpdate

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
