package module

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	modulev1 "cosmossdk.io/api/cosmos/group/module/v1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/genesis"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/group/simulation"
)

// ConsensusVersion defines the current x/group module consensus version.
const ConsensusVersion = 3

var (
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

type AppModule struct {
	AppModuleBasic
	keeper     keeper.Keeper
	bankKeeper group.BankKeeper
	accKeeper  group.AccountKeeper
	registry   cdctypes.InterfaceRegistry
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak group.AccountKeeper, bk group.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc, ac: ak, genesisHandler: keeper.HasGenesis},
		keeper:         keeper,
		bankKeeper:     bk,
		accKeeper:      ak,
		registry:       registry,
	}
}

var (
	_ appmodule.AppModule     = AppModule{}
	_ appmodule.HasEndBlocker = AppModule{}
)

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

type AppModuleBasic struct {
	cdc            codec.Codec
	ac             address.Codec
	genesisHandler appmodule.HasGenesis
}

// Name returns the group module's name.
func (AppModuleBasic) Name() string {
	return group.ModuleName
}

// DefaultGenesis returns default genesis state as raw bytes for the group
// module.
func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	target := genesis.RawJSONTarget{}
	if err := a.genesisHandler.DefaultGenesis(target.Target()); err != nil {
		panic(err)
	}

	defaultJSON, err := target.JSON()
	if err != nil {
		panic(err)
	}

	return defaultJSON
}

// ValidateGenesis performs genesis state validation for the group module.
func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	source, err := genesis.SourceFromRawJSON(bz)
	if err != nil {
		panic(err)
	}

	return a.genesisHandler.ValidateGenesis(source)
}

// GetQueryCmd returns the cli query commands for the group module
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.QueryCmd(a.Name())
}

// GetTxCmd returns the transaction commands for the group module
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.TxCmd(a.Name(), a.ac)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the group module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *gwruntime.ServeMux) {
	if err := group.RegisterQueryHandlerClient(context.Background(), mux, group.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// RegisterInterfaces registers the group module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	group.RegisterInterfaces(registry)
}

// RegisterLegacyAminoCodec registers the group module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	group.RegisterLegacyAminoCodec(cdc)
}

// Name returns the group module's name.
func (AppModule) Name() string {
	return group.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper)
}

// InitGenesis performs genesis initialization for the group module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	source, err := genesis.SourceFromRawJSON(data)
	if err != nil {
		panic(err)
	}

	if err := am.keeper.InitGenesis(ctx, source); err != nil {
		panic(err)
	}

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the group
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	target := genesis.RawJSONTarget{}
	if err := am.keeper.ExportGenesis(ctx, target.Target()); err != nil {
		panic(err)
	}

	exported, err := target.JSON()
	if err != nil {
		panic(err)
	}

	return exported
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	group.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	group.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper)
	if err := cfg.RegisterMigration(group.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", group.ModuleName, err))
	}

	if err := cfg.RegisterMigration(group.ModuleName, 2, m.Migrate2to3); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 2 to 3: %v", group.ModuleName, err))
	}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// EndBlock implements the group module's EndBlock.
func (am AppModule) EndBlock(ctx context.Context) error {
	c := sdk.UnwrapSDKContext(ctx)
	return EndBlocker(c, am.keeper)
}

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the group module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// RegisterStoreDecoder registers a decoder for group module's types
func (am AppModule) RegisterStoreDecoder(sdr simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		am.registry,
		simState.AppParams, simState.Cdc, simState.TxConfig,
		am.accKeeper, am.bankKeeper, am.keeper, am.cdc,
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

type GroupInputs struct {
	depinject.In

	Config           *modulev1.Module
	KVStoreService   store.KVStoreService
	Cdc              codec.Codec
	AccountKeeper    group.AccountKeeper
	BankKeeper       group.BankKeeper
	Registry         cdctypes.InterfaceRegistry
	MsgServiceRouter baseapp.MessageRouter
}

type GroupOutputs struct {
	depinject.Out

	GroupKeeper keeper.Keeper
	Module      appmodule.AppModule
}

func ProvideModule(in GroupInputs) GroupOutputs {
	/*
		Example of setting group params:
		in.Config.MaxMetadataLen = 1000
		in.Config.MaxExecutionPeriod = "1209600s"
	*/

	k := keeper.NewKeeper(
		in.KVStoreService,
		in.MsgServiceRouter,
		in.AccountKeeper,
		group.Config{MaxExecutionPeriod: in.Config.MaxExecutionPeriod.AsDuration(), MaxMetadataLen: in.Config.MaxMetadataLen},
	)

	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.Registry)

	return GroupOutputs{GroupKeeper: k, Module: m}
}
