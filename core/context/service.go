package context

import (
	"context"
	"log"

	"cosmossdk.io/core/comet"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/header"
)

// ContextKey defines a type alias for a stdlib Context key.
type ContextKey string

// SdkContextKey is the key in the context.Context which holds the sdk.Context.
const SdkContextKey ContextKey = "sdk-context"

// ExecMode defines the execution mode which can be set on a Context.
type ExecMode uint8

// All possible execution modes.
const (
	ExecModeCheck ExecMode = iota
	ExecModeReCheck
	ExecModeSimulate
	ExecModePrepareProposal
	ExecModeProcessProposal
	ExecModeVoteExtension
	ExecModeFinalize
)

// Context is an interface that extends the stdlib context.Context with SDK-specific
// This context is specific to modules
type Context interface {
	Context() context.Context
	HeaderInfo() header.Info
	CometInfo() comet.BlockInfo
	Logger() log.Logger
	EventManager() event.Service
	ExecMode() ExecMode
}

func UnwrapSDKContext(ctx context.Context) Context {
	if ctx == nil {
		return nil
	}
	return ctx.Value(SdkContextKey).(Context)
}
