package upgrade

import (
	"context"
	"encoding/json"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	modulev1 "cosmossdk.io/api/cosmos/upgrade/module/v1"
	"github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
	"github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func init() {
	types.RegisterLegacyAminoCodec(codec.NewLegacyAmino())
}

const (
	consensusVersion uint64 = 2
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic implements the sdk.AppModuleBasic interface
type AppModuleBasic struct{}

// Name returns the ModuleName
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the upgrade types on the LegacyAmino codec
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the upgrade module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetQueryCmd returns the cli query commands for this module
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// GetTxCmd returns the transaction commands for this module
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (b AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements the sdk.AppModule interface
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// RegisterInvariants does nothing, there are no invariants to enforce
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Deprecated: Route returns the message routing key for the upgrade module.
func (AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the route we respond to for abci queries
func (AppModule) QuerierRoute() string { return types.QuerierKey }

// LegacyQuerierHandler registers a query handler to respond to the module-specific queries
func (am AppModule) LegacyQuerierHandler(legacyQuerierCdc *codec.LegacyAmino) sdk.Querier {
	return keeper.NewQuerier(am.keeper, legacyQuerierCdc)
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)

	m := keeper.NewMigrator(am.keeper)
	err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2)
	if err != nil {
		panic(err)
	}
}

// InitGenesis is ignored, no sense in serializing future upgrades
func (am AppModule) InitGenesis(_ sdk.Context, _ codec.JSONCodec, _ json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// DefaultGenesis is an empty object
func (AppModuleBasic) DefaultGenesis(_ codec.JSONCodec) json.RawMessage {
	return []byte("{}")
}

// ValidateGenesis is always successful, as we ignore the value
func (AppModuleBasic) ValidateGenesis(_ codec.JSONCodec, config client.TxEncodingConfig, _ json.RawMessage) error {
	return nil
}

// ExportGenesis is always empty, as InitGenesis does nothing either
func (am AppModule) ExportGenesis(_ sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	return am.DefaultGenesis(cdc)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return consensusVersion }

// BeginBlock calls the upgrade module hooks
//
// CONTRACT: this is registered in BeginBlocker *before* all other modules' BeginBlock functions
func (am AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(am.keeper, ctx, req)
}

//
// New App Wiring Setup
//

func init() {
	appmodule.Register(&modulev1.Module{},
		appmodule.Provide(provideModuleBasic, provideModule),
	)
}

func provideModuleBasic() runtime.AppModuleBasicWrapper {
	return runtime.WrapAppModuleBasic(AppModuleBasic{})
}

type upgradeInputs struct {
	depinject.In

	ModuleKey depinject.OwnModuleKey
	Config    *modulev1.Module
	Key       *store.KVStoreKey
	Cdc       codec.Codec

	AppOpts   servertypes.AppOptions    `optional:"true"`
	Authority map[string]sdk.AccAddress `optional:"true"`
}

type upgradeOutputs struct {
	depinject.Out

	UpgradeKeeper keeper.Keeper
	Module        runtime.AppModuleWrapper
	GovHandler    govv1beta1.HandlerRoute
}

func provideModule(in upgradeInputs) upgradeOutputs {
	var (
		homePath           string
		skipUpgradeHeights = make(map[int64]bool)
	)

	if in.AppOpts != nil {
		for _, h := range cast.ToIntSlice(in.AppOpts.Get(server.FlagUnsafeSkipUpgrades)) {
			skipUpgradeHeights[int64(h)] = true
		}

		homePath = cast.ToString(in.AppOpts.Get(flags.FlagHome))
	}

	authority, ok := in.Authority[depinject.ModuleKey(in.ModuleKey).Name()]
	if !ok {
		// default to governance authority if not provided
		authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	}

	// set the governance module account as the authority for conducting upgrades
	k := keeper.NewKeeper(skipUpgradeHeights, in.Key, in.Cdc, homePath, nil, authority.String())
	m := NewAppModule(k)
	gh := govv1beta1.HandlerRoute{RouteKey: types.RouterKey, Handler: NewSoftwareUpgradeProposalHandler(k)}

	return upgradeOutputs{UpgradeKeeper: k, Module: runtime.WrapAppModule(m), GovHandler: gh}
}
