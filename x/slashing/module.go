package slashing

import (
	"context"
	"encoding/json"
	"fmt"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"

	modulev1 "cosmossdk.io/api/cosmos/slashing/module/v1"
	"cosmossdk.io/core/appmodule"
	store "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/slashing/keeper"
	"cosmossdk.io/x/slashing/simulation"
	"cosmossdk.io/x/slashing/types"
	staking "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ConsensusVersion defines the current x/slashing module consensus version.
const ConsensusVersion = 4

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}

	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
)

// AppModuleBasic defines the basic application module used by the slashing module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the slashing module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the slashing module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the module's interface types
func (b AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the slashing
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the slashing module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return types.ValidateGenesis(data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the slashig module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// AppModule implements an application module for the slashing module.
type AppModule struct {
	AppModuleBasic

	registry cdctypes.InterfaceRegistry

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
	stakingKeeper types.StakingKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper keeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	sk types.StakingKeeper,
	registry cdctypes.InterfaceRegistry,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		stakingKeeper:  sk,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))

	m := keeper.NewMigrator(am.keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", types.ModuleName, err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", types.ModuleName, err))
	}
	if err := cfg.RegisterMigration(types.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 3 to 4: %v", types.ModuleName, err))
	}
}

// InitGenesis performs genesis initialization for the slashing module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, am.stakingKeeper, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the slashing
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// BeginBlock returns the begin blocker for the slashing module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return BeginBlocker(ctx, am.keeper)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the slashing module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for slashing module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the slashing module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		am.registry, simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accountKeeper, am.bankKeeper, am.keeper, am.stakingKeeper,
	)
}

//
// App Wiring Setup
//

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	StoreService store.KVStoreService
	Cdc          codec.Codec
	LegacyAmino  *codec.LegacyAmino
	Registry     cdctypes.InterfaceRegistry

	AccountKeeper types.AccountKeeper
	BankKeeper    types.BankKeeper
	StakingKeeper types.StakingKeeper
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
	Hooks  staking.StakingHooksWrapper
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authStr, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(fmt.Errorf("unable to decode authority in slashing: %w", err))
	}

	k := keeper.NewKeeper(in.Cdc, in.LegacyAmino, in.StoreService, in.StakingKeeper, authStr)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.StakingKeeper, in.Registry)
	return ModuleOutputs{
		Keeper: k,
		Module: m,
		Hooks:  staking.StakingHooksWrapper{StakingHooks: k.Hooks()},
	}
}
