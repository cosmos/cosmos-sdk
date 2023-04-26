package module

import (
	"context"

	"cosmossdk.io/core/appmodule"
)

var _ appmodule.HasGenesis = AppModule{}

// DefaultGenesis writes the default group genesis for this module to the target.
func (am AppModule) DefaultGenesis(genesisTarget appmodule.GenesisTarget) error {
	return am.keeper.DefaultGenesis(genesisTarget)
}

// ValidateGenesis validates the group module genesis data read from the source.
func (am AppModule) ValidateGenesis(genesisSource appmodule.GenesisSource) error {
	return am.keeper.ValidateGenesis(genesisSource)
}

// InitGenesis initializes the group state from the genesis source.
func (am AppModule) InitGenesis(ctx context.Context, genesisSource appmodule.GenesisSource) error {
	return am.keeper.InitGenesis(ctx, genesisSource)
}

// ExportGenesis exports the group state to the genesis target.
func (am AppModule) ExportGenesis(ctx context.Context, genesisTarget appmodule.GenesisTarget) error {
	return am.keeper.ExportGenesis(ctx, genesisTarget)
}
