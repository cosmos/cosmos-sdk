package module

import (
	"context"
	"encoding/json"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	grpc "google.golang.org/grpc"

	modulev1 "cosmossdk.io/api/cosmos/benchmark/module/v1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/log"
	"cosmossdk.io/x/benchmark"
	gen "cosmossdk.io/x/benchmark/generator"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var (
	_ module.HasGRPCGateway = &AppModule{}
	_ appmodule.AppModule   = &AppModule{}
	_ appmodule.HasGenesis  = &AppModule{}
)

type AppModule struct {
	keeper        *Keeper
	storeKeys     []string
	genesisParams *modulev1.GenesisParams
	log           log.Logger
}

func NewAppModule(
	genesisParams *modulev1.GenesisParams,
	storeKeys []string,
	kvMap KVServiceMap,
	logger log.Logger,
) *AppModule {
	return &AppModule{
		keeper:    NewKeeper(kvMap),
		storeKeys: storeKeys,
		log:       logger,
	}
}

// DefaultGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) DefaultGenesis() json.RawMessage { return nil }

// ExportGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) ExportGenesis(context.Context) (json.RawMessage, error) { return nil, nil }

// InitGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) InitGenesis(ctx context.Context, _ json.RawMessage) error {
	g := gen.NewGenerator(gen.Options{
		Seed:        a.genesisParams.Seed,
		KeyMean:     a.genesisParams.KeyMeanLength,
		KeyStdDev:   a.genesisParams.KeyStdDevLength,
		ValueMean:   a.genesisParams.ValueMeanLength,
		ValueStdDev: a.genesisParams.ValueStdDevLength,
	})
	skCount := uint64(len(a.storeKeys))
	for i := range a.genesisParams.KeyCount {
		if i%100_000 == 0 {
			a.log.Warn("init genesis", "progress", i, "total", a.genesisParams.KeyCount)
		}
		sk := a.storeKeys[g.UintN(skCount)]
		a.keeper.set(ctx, sk, g.Key(), g.Value())
	}
	return nil
}

// ValidateGenesis implements appmodulev2.HasGenesis.
func (a *AppModule) ValidateGenesis(data json.RawMessage) error { return nil }

func (a *AppModule) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {
}

// RegisterServices registers module services.
func (am *AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	benchmark.RegisterMsgServer(registrar, am.keeper)
	return nil
}

func (a *AppModule) IsOnePerModuleType() {}

func (a *AppModule) IsAppModule() {}
