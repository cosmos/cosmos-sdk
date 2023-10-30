package counter

import (
	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	modulev1 "cosmossdk.io/api/cosmos/counter/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/event"
	storetypes "cosmossdk.io/core/store"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/counter/keeper"
	"github.com/cosmos/cosmos-sdk/x/counter/types"
)

var (
	_ module.AppModuleBasic = AppModule{}

	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}
)

// AppModuleBasic defines the basic application module used by the consensus module.
type AppModuleBasic struct{}

// Name returns the consensus module's name.
func (AppModuleBasic) Name() string { return types.ModuleName }

// RegisterLegacyAminoCodec registers the consensus module's types on the LegacyAmino codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {}

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// AppModule implements an application module
type AppModule struct {
	AppModuleBasic

	keeper keeper.Keeper
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

func init() {
	appmodule.Register(
		&modulev1.Module{},
		appmodule.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *modulev1.Module
	StoreService storetypes.KVStoreService
	EventManager event.Service
}

type ModuleOutputs struct {
	depinject.Out

	Keeper keeper.Keeper
	Module appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	k := keeper.NewKeeper(in.StoreService, in.EventManager)
	m := NewAppModule(k)

	return ModuleOutputs{
		Keeper: k,
		Module: m,
	}
}
