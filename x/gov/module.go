package gov

// DONTCOVER

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	abci "github.com/cometbft/cometbft/abci/types"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	modulev1 "cosmossdk.io/api/cosmos/gov/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/simulation"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const ConsensusVersion = 4

var (
	_ module.EndBlockAppModule   = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the gov module.
type AppModuleBasic struct {
	cdc                    codec.Codec
	legacyProposalHandlers []govclient.ProposalHandler // legacy proposal handlers which live in governance cli and rest
}

// NewAppModuleBasic creates a new AppModuleBasic object
func NewAppModuleBasic(legacyProposalHandlers []govclient.ProposalHandler) AppModuleBasic {
	return AppModuleBasic{
		legacyProposalHandlers: legacyProposalHandlers,
	}
}

// Name returns the gov module's name.
func (AppModuleBasic) Name() string {
	return govtypes.ModuleName
}

// RegisterLegacyAminoCodec registers the gov module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	v1beta1.RegisterLegacyAminoCodec(cdc)
	v1.RegisterLegacyAminoCodec(cdc)
}

// DefaultGenesis returns default genesis state as raw bytes for the gov
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(v1.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the gov module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var data v1.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", govtypes.ModuleName, err)
	}

	return v1.ValidateGenesis(&data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the gov module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := v1.RegisterQueryHandlerClient(context.Background(), mux, v1.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
	if err := v1beta1.RegisterQueryHandlerClient(context.Background(), mux, v1beta1.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command for the gov module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	legacyProposalCLIHandlers := getProposalCLIHandlers(a.legacyProposalHandlers)

	return cli.NewTxCmd(legacyProposalCLIHandlers)
}

func getProposalCLIHandlers(handlers []govclient.ProposalHandler) []*cobra.Command {
	proposalCLIHandlers := make([]*cobra.Command, 0, len(handlers))
	for _, proposalHandler := range handlers {
		proposalCLIHandlers = append(proposalCLIHandlers, proposalHandler.CLIHandler())
	}
	return proposalCLIHandlers
}

// GetQueryCmd returns the root query command for the gov module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// RegisterInterfaces implements InterfaceModule.RegisterInterfaces
func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	v1.RegisterInterfaces(registry)
	v1beta1.RegisterInterfaces(registry)
}

// AppModule implements an application module for the gov module.
type AppModule struct {
	AppModuleBasic

	keeper        *keeper.Keeper
	accountKeeper govtypes.AccountKeeper
	bankKeeper    govtypes.BankKeeper

	// legacySubspace is used solely for migration of x/params managed parameters
	legacySubspace govtypes.ParamSubspace
}

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec, keeper *keeper.Keeper,
	ak govtypes.AccountKeeper, bk govtypes.BankKeeper, ss govtypes.ParamSubspace,
) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		legacySubspace: ss,
	}
}

var _ appmodule.AppModule = AppModule{}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule, ProvideKeyTable),
		appmodule.Invoke(InvokeAddRoutes, InvokeSetHooks))
}

type GovInputs struct {
	depinject.In

	Config           *modulev1.Module
	Cdc              codec.Codec
	Key              *store.KVStoreKey
	ModuleKey        depinject.OwnModuleKey
	MsgServiceRouter *baseapp.MsgServiceRouter

	AccountKeeper govtypes.AccountKeeper
	BankKeeper    govtypes.BankKeeper
	StakingKeeper govtypes.StakingKeeper

	// LegacySubspace is used solely for migration of x/params managed parameters
	LegacySubspace govtypes.ParamSubspace
}

type GovOutputs struct {
	depinject.Out

	Module       appmodule.AppModule
	Keeper       *keeper.Keeper
	HandlerRoute v1beta1.HandlerRoute
}

func ProvideModule(in GovInputs) GovOutputs {
	kConfig := govtypes.DefaultConfig()
	if in.Config.MaxMetadataLen != 0 {
		kConfig.MaxMetadataLen = in.Config.MaxMetadataLen
	}

	// default to governance authority if not provided
	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}

	k := keeper.NewKeeper(
		in.Cdc,
		in.Key,
		in.AccountKeeper,
		in.BankKeeper,
		in.StakingKeeper,
		in.MsgServiceRouter,
		kConfig,
		authority.String(),
	)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.LegacySubspace)
	hr := v1beta1.HandlerRoute{Handler: v1beta1.ProposalHandler, RouteKey: govtypes.RouterKey}

	return GovOutputs{Module: m, Keeper: k, HandlerRoute: hr}
}

func ProvideKeyTable() paramtypes.KeyTable {
	return v1.ParamKeyTable() //nolint:staticcheck
}

func InvokeAddRoutes(keeper *keeper.Keeper, routes []v1beta1.HandlerRoute) {
	if keeper == nil || routes == nil {
		return
	}

	// Default route order is a lexical sort by RouteKey.
	// Explicit ordering can be added to the module config if required.
	slices.SortFunc(routes, func(x, y v1beta1.HandlerRoute) bool {
		return x.RouteKey < y.RouteKey
	})

	router := v1beta1.NewRouter()
	for _, r := range routes {
		router.AddRoute(r.RouteKey, r.Handler)
	}
	keeper.SetLegacyRouter(router)
}

func InvokeSetHooks(keeper *keeper.Keeper, govHooks map[string]govtypes.GovHooksWrapper) error {
	if keeper == nil || govHooks == nil {
		return nil
	}

	// Default ordering is lexical by module name.
	// Explicit ordering can be added to the module config if required.
	modNames := maps.Keys(govHooks)
	order := modNames
	sort.Strings(order)

	var multiHooks govtypes.MultiGovHooks
	for _, modName := range order {
		hook, ok := govHooks[modName]
		if !ok {
			return fmt.Errorf("can't find staking hooks for module %s", modName)
		}
		multiHooks = append(multiHooks, hook)
	}

	keeper.SetHooks(multiHooks)
	return nil
}

// Name returns the gov module's name.
func (AppModule) Name() string {
	return govtypes.ModuleName
}

// RegisterInvariants registers module invariants
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, am.keeper, am.bankKeeper)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	msgServer := keeper.NewMsgServerImpl(am.keeper)
	v1beta1.RegisterMsgServer(cfg.MsgServer(), keeper.NewLegacyMsgServerImpl(am.accountKeeper.GetModuleAddress(govtypes.ModuleName).String(), msgServer))
	v1.RegisterMsgServer(cfg.MsgServer(), msgServer)

	legacyQueryServer := keeper.NewLegacyQueryServer(am.keeper)
	v1beta1.RegisterQueryServer(cfg.QueryServer(), legacyQueryServer)
	v1.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper, am.legacySubspace)
	err := cfg.RegisterMigration(govtypes.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 1 to 2: %v", err))
	}
	err = cfg.RegisterMigration(govtypes.ModuleName, 2, m.Migrate2to3)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 2 to 3: %v", err))
	}
	err = cfg.RegisterMigration(govtypes.ModuleName, 3, m.Migrate3to4)
	if err != nil {
		panic(fmt.Sprintf("failed to migrate x/gov from version 3 to 4: %v", err))
	}
}

// InitGenesis performs genesis initialization for the gov module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState v1.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	InitGenesis(ctx, am.accountKeeper, am.bankKeeper, am.keeper, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the gov
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// EndBlock returns the end blocker for the gov module. It returns no validator
// updates.
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, am.keeper)
	return []abci.ValidatorUpdate{}
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the gov module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the gov content functions used to
// simulate governance proposals.
func (AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent { //nolint:staticcheck
	return simulation.ProposalContents()
}

// ProposalMsgs returns all the gov msgs used to simulate governance proposals.
func (AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return simulation.ProposalMsgs()
}

// RegisterStoreDecoder registers a decoder for gov module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[govtypes.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc,
		am.accountKeeper, am.bankKeeper, am.keeper,
		simState.ProposalMsgs, simState.LegacyProposalContents,
	)
}
