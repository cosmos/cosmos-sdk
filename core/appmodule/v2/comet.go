package appmodule

import "context"

// HasPreBlocker is the extension interface that modules should implement to run
// custom logic before BeginBlock.
type HasPreBlocker interface {
	AppModule
	// PreBlock is method that will be run before BeginBlock.
	PreBlock(context.Context) error
}
