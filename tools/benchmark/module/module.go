package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	_ "cosmossdk.io/api/cosmos/benchmark/v1" // for some reason this is required to make msg server registration work
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/tools/benchmark"
	"cosmossdk.io/tools/benchmark/client/cli"
	gen "cosmossdk.io/tools/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	_ module.HasGRPCGateway = &AppModule{}
	_ appmodule.AppModule   = &AppModule{}
	_ appmodule.HasGenesis  = &AppModule{}
)

type AppModule struct {
	keeper        *Keeper
	storeKeys     []string
	genesisParams *modulev1.GeneratorParams
	log           log.Logger
}

func NewAppModule(
	genesisParams *modulev1.GeneratorParams,
	storeKeys []string,
	kvMap KVServiceMap,
	logger log.Logger,
) *AppModule {
	return &AppModule{
		genesisParams: genesisParams,
		keeper:        NewKeeper(kvMap),
		storeKeys:     storeKeys,
		log:           logger,
	}
}

// DefaultGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) DefaultGenesis() json.RawMessage {
	return nil
}

// ExportGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) ExportGenesis(context.Context) (json.RawMessage, error) { return nil, nil }

// InitGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) InitGenesis(ctx context.Context, _ json.RawMessage) error {
	a.genesisParams.BucketCount = uint64(len(a.storeKeys))
	g := gen.NewGenerator(gen.Options{GeneratorParams: a.genesisParams})
	i := 0
	for kv := range g.GenesisSet() {
		i++
		if i%100_000 == 0 {
			a.log.Warn("benchmark: init genesis", "progress", i, "total", a.genesisParams.GenesisCount)
		}
		sk := a.storeKeys[kv.StoreKey]
		key := gen.Bytes(kv.Key.Seed(), kv.Key.Length())
		value := gen.Bytes(kv.Value.Seed(), kv.Value.Length())
		err := a.keeper.set(ctx, sk, key, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) ValidateGenesis(data json.RawMessage) error { return nil }

func (a *AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
}

// RegisterServices registers module services.
func (a *AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	benchmark.RegisterMsgServer(registrar, a.keeper)
	return nil
}

func (a *AppModule) RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations(
		(*transaction.Msg)(nil),
		&benchmark.MsgLoadTest{})
	msgservice.RegisterMsgServiceDesc(registrar, &benchmark.Msg_serviceDesc)
}

func (a *AppModule) GetTxCmd() *cobra.Command {
	return cli.NewTxCmd(a.genesisParams)
}

func (a *AppModule) IsOnePerModuleType() {}

func (a *AppModule) IsAppModule() {}
