package module

import (
	modulev1 "cosmossdk.io/api/cosmos/feegrant/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"cosmossdk.io/x/feegrant"
	"cosmossdk.io/x/feegrant/keeper"
	"cosmossdk.io/x/feegrant/simulation"

	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

var _ depinject.OnePerModuleType = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.RegisterModule(&modulev1.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type FeegrantInputs struct {
	depinject.In

	Environment   appmodule.Environment
	Cdc           codec.Codec
	AccountKeeper feegrant.AccountKeeper
	BankKeeper    feegrant.BankKeeper
	Registry      cdctypes.InterfaceRegistry
}

func ProvideModule(in FeegrantInputs) (keeper.Keeper, appmodule.AppModule) {
	k := keeper.NewKeeper(in.Environment, in.Cdc, in.AccountKeeper)
	m := NewAppModule(in.Cdc, in.AccountKeeper, in.BankKeeper, k, in.Registry)
	return k, m
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the feegrant module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for feegrant module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[feegrant.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns all the feegrant module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		am.registry, simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accountKeeper, am.bankKeeper, am.keeper, am.accountKeeper.AddressCodec(),
	)
}
