package authz

import (
	"context"
	"encoding/json"
	"math/rand"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	modulev1 "cosmossdk.io/api/cosmos/authz/module/v1"
	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/depinject"
	"github.com/cosmos/cosmos-sdk/runtime"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/authz/client/cli"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/simulation"
)

var (
	_ module.AppModule           = AppModule{}
	_ module.AppModuleBasic      = AppModuleBasic{}
	_ module.AppModuleSimulation = AppModule{}
)

// AppModuleBasic defines the basic application module used by the authz module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the authz module's name.
func (AppModuleBasic) Name() string {
	return authz.ModuleName
}

// RegisterServices registers a gRPC query service to respond to the
// module-specific gRPC queries.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	authz.RegisterQueryServer(cfg.QueryServer(), am.keeper)
	authz.RegisterMsgServer(cfg.MsgServer(), am.keeper)
	m := keeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(authz.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(err)
	}
}

// RegisterLegacyAminoCodec registers the authz module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	authz.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers the authz module's interface types
func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	authz.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the authz
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(authz.DefaultGenesisState())
}

// ValidateGenesis performs genesis state validation for the authz module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config sdkclient.TxEncodingConfig, bz json.RawMessage) error {
	var data authz.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return sdkerrors.Wrapf(err, "failed to unmarshal %s genesis state", authz.ModuleName)
	}

	return authz.ValidateGenesis(data)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the authz module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx sdkclient.Context, mux *gwruntime.ServeMux) {
	if err := authz.RegisterQueryHandlerClient(context.Background(), mux, authz.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetQueryCmd returns the cli query commands for the authz module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// GetTxCmd returns the transaction commands for the authz module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper        keeper.Keeper
	accountKeeper authz.AccountKeeper
	bankKeeper    authz.BankKeeper
	registry      cdctypes.InterfaceRegistry
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak authz.AccountKeeper, bk authz.BankKeeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		accountKeeper:  ak,
		bankKeeper:     bk,
		registry:       registry,
	}
}

// Name returns the authz module's name.
func (AppModule) Name() string {
	return authz.ModuleName
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the authz module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

func (am AppModule) NewHandler() sdk.Handler {
	return nil
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return "" }

// LegacyQuerierHandler returns the authz module sdk.Querier.
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return nil
}

// InitGenesis performs genesis initialization for the authz module. It returns
// no validator updates.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState authz.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, &genesisState)
	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the authz
// module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 2 }

// BeginBlock returns the begin blocker for the authz module.
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, am.keeper)
}

// EndBlock does nothing
func (am AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(
			provideModuleBasic,
			provideModule,
		),
	)
}

func provideModuleBasic() runtime.AppModuleBasicWrapper {
	return runtime.WrapAppModuleBasic(AppModuleBasic{})
}

type authzInputs struct {
	depinject.In

	Key              *store.KVStoreKey
	Cdc              codec.Codec
	AccountKeeper    authz.AccountKeeper `key:"cosmos.auth.v1.AccountKeeper"`
	BankKeeper       authz.BankKeeper    `key:"cosmos.bank.v1.Keeper"`
	Registry         cdctypes.InterfaceRegistry
	MsgServiceRouter *baseapp.MsgServiceRouter
}

func provideModule(in authzInputs) (keeper.Keeper, runtime.AppModuleWrapper) {
	k := keeper.NewKeeper(in.Key, in.Cdc, in.MsgServiceRouter, in.AccountKeeper)
	m := NewAppModule(in.Cdc, k, in.AccountKeeper, in.BankKeeper, in.Registry)
	return k, runtime.WrapAppModule(m)
}

// ____________________________________________________________________________

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the authz module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	simulation.RandomizedGenState(simState)
}

// ProposalContents returns all the authz content functions used to
// simulate governance proposals.
func (am AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized authz param changes for the simulator.
func (AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return nil
}

// RegisterStoreDecoder registers a decoder for authz module's types
func (am AppModule) RegisterStoreDecoder(sdr sdk.StoreDecoderRegistry) {
	sdr[keeper.StoreKey] = simulation.NewDecodeStore(am.cdc)
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc,
		am.accountKeeper, am.bankKeeper, am.keeper, am.cdc,
	)
}
