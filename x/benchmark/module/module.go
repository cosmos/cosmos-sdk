package module

import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpc "google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/x/benchmark"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.HasGRPCGateway = AppModule{}
	_ appmodule.AppModule   = AppModule{}
)

type AppModule struct {
	keeper *Keeper
}

func NewAppModule(collector *KVServiceCollector) AppModule {
	return AppModule{
		keeper: NewKeeper(collector),
	}
}

func (a AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	benchmark.RegisterMsgServer(registrar, am.keeper)
	return nil
}

func (a AppModule) IsOnePerModuleType() {}

func (a AppModule) IsAppModule() {}
