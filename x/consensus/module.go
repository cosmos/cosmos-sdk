package consensus

import (
	"context"
	"encoding/json"

	modulev1 "cosmossdk.io/api/cosmos/consensus/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	store "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	"github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// ConsensusVersion defines the current x/bank module consensus version.
const ConsensusVersion = 1

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module used by the consensus_param module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the consensus_param module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the consensus_param module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// DefaultGenesis returns default genesis state as raw bytes for the consensus_param
// module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	// nil is returned since default genesis of consensus params is handled by tendermint
	return nil
}

// ValidateGenesis performs genesis state validation
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	return nil
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// GetTxCmd returns the root tx command
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil
}

// GetQueryCmd returns no root query command
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQuerier(am.keeper))
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// Name returns the consensus_param module's name.
func (AppModule) Name() string { return types.ModuleName }

// InitGenesis is handled by for init genesis of consensus_param
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate {
	// nil is returned since initgenesis of consensus params is handled by tendermint
	return nil
}

// ExportGenesis is handled by tendermint export of genesis
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// nil is returned since ExportGenesis of consensus params is handled by tendermint and baseapp
	return nil
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

// RegisterInvariants does nothing, there are no invariants to enforce
func (am AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {}

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(provideModuleBasic, provideModule),
	)
}

func provideModuleBasic() runtime.AppModuleBasicWrapper {
	return runtime.WrapAppModuleBasic(AppModuleBasic{})
}

type consensusParamInputs struct {
	depinject.In

	Cdc       codec.Codec
	Key       *store.KVStoreKey
	ModuleKey depinject.OwnModuleKey
	Authority map[string]sdk.AccAddress `optional:"true"`
}

type consensusParamOutputs struct {
	depinject.Out

	Keeper        keeper.Keeper
	Module        runtime.AppModuleWrapper
	BaseAppOption runtime.BaseAppOption
}

func provideModule(in consensusParamInputs) consensusParamOutputs {
	authority, ok := in.Authority[depinject.ModuleKey(in.ModuleKey).Name()]
	if !ok {
		// default to governance authority if not provided
		authority = authtypes.NewModuleAddress(govtypes.ModuleName)
	}

	k := keeper.NewKeeper(in.Cdc, in.Key, authority.String())
	m := NewAppModule(in.Cdc, k)
	baseappOpt := func(app *baseapp.BaseApp) {
		app.SetParamStore(&k)
	}

	return consensusParamOutputs{
		Keeper:        k,
		Module:        runtime.WrapAppModule(m),
		BaseAppOption: baseappOpt,
	}
}
