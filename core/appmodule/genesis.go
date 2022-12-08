package appmodule

import (
	"context"

	"cosmossdk.io/core/genesis"
)

// HasGenesis is the extension interface that modules should implement to handle
// genesis data and state initialization.
type HasGenesis interface {
	AppModule

	// DefaultGenesis writes the default genesis for this module to the target.
	DefaultGenesis(genesis.Target) error

	// ValidateGenesis validates the genesis data read from the source.
	ValidateGenesis(genesis.Source) error

	// InitGenesis initializes module state from the genesis source.
	InitGenesis(context.Context, genesis.Source) error

	// ExportGenesis exports module state to the genesis target.
	ExportGenesis(context.Context, genesis.Target) error
}
