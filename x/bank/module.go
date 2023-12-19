package bank

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	modulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/core/appmodule"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/bank/client/cli"
	"cosmossdk.io/x/bank/keeper"
	"cosmossdk.io/x/bank/simulation"
	"cosmossdk.io/x/bank/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ConsensusVersion defines the current x/bank module consensus version.
const ConsensusVersion = 4

var (
	_ module.AppModuleBasic      = AppModule{}
	_ module.AppModuleSimulation = AppModule{}
	_ module.HasGenesis          = AppModule{}
	_ module.HasServices         = AppModule{}
	_ module.HasInvariants       = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// AppModuleBasic defines the basic application module used by the bank module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the bank module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the bank module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the bank
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the bank module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}

	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the bank module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the bank module.
func (ab AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd()
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module for the bank module.
type AppModule struct {
	AppModuleBasic

	keeper        keeper.Keeper
	accountKeeper types.AccountKeeper
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper.(keeper.BaseKeeper))

	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 1 to 2: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 2 to 3: %v", err))
	}

	if err := cfg.RegisterMigration(types.ModuleName, 3, m.Migrate3to4); err != nil {
		panic(fmt.Sprintf("failed to migrate x/bank from version 3 to 4: %v", err))
	}
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, accountKeeper types.AccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  accountKeeper,
	}
}

// RegisterInvariants registers the bank module invariants.
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// QuerierRoute returns the bank module's querier route name.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// InitGenesis performs genesis initialization for the bank module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx context.Context, cdc codec.JSONCodec, data json.RawMessage) {
	start := time.Now()
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	telemetry.MeasureSince(start, "InitGenesis", "crisis", "unmarshal")

	am.keeper.InitGenesis(ctx, &genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the bank
// module.
func (am AppModule) ExportGenesis(ctx context.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the bank module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for supply module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {
	sdr[types.StoreKey] = simtypes.NewStoreDecoderFuncFromCollectionsSchema(am.keeper.(keeper.BaseKeeper).Schema)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc, simState.TxConfig, am.accountKeeper, am.keeper,
	)
}

// App Wiring Setup

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	Cdc          codec.Codec
	StoreService corestore.KVStoreService
	Logger       log.Logger

	AccountKeeper types.AccountKeeper
}

type ModuleOutputs struct {
	depinject.Out

	BankKeeper keeper.BaseKeeper
	Module     appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	// Configure blocked module accounts.
	//
	// Default behavior for blockedAddresses is to regard any module mentioned in
	// AccountKeeper's module account permissions as blocked.
	blockedAddresses := make(map[string]bool)
	if len(in.Config.BlockedModuleAccountsOverride) > 0 {
		for _, moduleName := range in.Config.BlockedModuleAccountsOverride {
			addrStr, err := in.AccountKeeper.AddressCodec().BytesToString(authtypes.NewModuleAddress(moduleName))
			if err != nil {
				panic(err)
			}
			blockedAddresses[addrStr] = true
		}
	} else {
		for _, permission := range in.AccountKeeper.GetModulePermissions() {
			addrStr, err := in.AccountKeeper.AddressCodec().BytesToString(permission.GetAddress())
			if err != nil {
				panic(err)
			}
			blockedAddresses[addrStr] = true
		}
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	authStr, err := in.AccountKeeper.AddressCodec().BytesToString(authority)
	if err != nil {
		panic(err)
	}

	bankKeeper := keeper.NewBaseKeeper(
		in.Cdc,
		in.StoreService,
		in.AccountKeeper,
		blockedAddresses,
		authStr,
		in.Logger,
	)
	m := NewAppModule(in.Cdc, bankKeeper, in.AccountKeeper)

	return ModuleOutputs{BankKeeper: bankKeeper, Module: m}
}
