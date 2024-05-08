package circuit

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/circuit/keeper"
	"cosmossdk.io/x/circuit/types"
)

// ConsensusVersion defines the current circuit module consensus version.
const ConsensusVersion = 1

var (
	_ appmodule.AppModule  = AppModule{}
	_ appmodule.HasGenesis = AppModule{}
)

// AppModule implements an application module for the circuit module.
type AppModule struct {
	keeper keeper.Keeper
}

func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) DefaultGenesis() json.RawMessage {
	// TODO implement me
	panic("implement me")
}

func (am AppModule) ValidateGenesis(data json.RawMessage) error {
	// TODO implement me
	panic("implement me")
}

func (am AppModule) InitGenesis(ctx context.Context, data json.RawMessage) error {
	// TODO implement me
	panic("implement me")
}

func (am AppModule) ExportGenesis(ctx context.Context) (json.RawMessage, error) {
	// TODO implement me
	panic("implement me")
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the circuit module's name.
func (AppModule) Name() string { return types.ModuleName }

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }
