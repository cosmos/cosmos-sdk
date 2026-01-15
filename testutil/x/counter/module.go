package counter

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/app"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter/keeper"
	"github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

var (
	_ appmodule.AppModule = AppModule{}
	_ app.Module          = AppModule{}
)

// AppModule implements an application module
type AppModule struct {
	keeper *keeper.Keeper
}

func NewExtendedAppModule() AppModule {
	k := keeper.NewExtendedKeeper()
	return NewAppModule(k)
}

func (am AppModule) StoreKeys() map[string]*storetypes.KVStoreKey {
	return am.keeper.StoreKeys()
}

func (am AppModule) ModuleAccountPermissions() map[string][]string {
	return nil
}

func (am AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {
}

func (am AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {
}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

// NewAppModule creates a new AppModule object
func NewAppModule(keeper *keeper.Keeper) AppModule {
	return AppModule{
		keeper: keeper,
	}
}

// ConsensusVersion implements HasConsensusVersion
func (AppModule) ConsensusVersion() uint64 { return 1 }

// Name returns the module's name.
func (AppModule) Name() string { return types.ModuleName }

// RegisterInterfaces registers interfaces and implementations of the bank module.
func (AppModule) RegisterInterfaces(registrar codectypes.InterfaceRegistry) {
	types.RegisterInterfaces(registrar)
}
