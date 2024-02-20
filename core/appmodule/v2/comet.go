package appmodule

import "context"

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
