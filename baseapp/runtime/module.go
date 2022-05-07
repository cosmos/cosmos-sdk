package runtime

import (
	"fmt"
	"io"

	"github.com/gogo/protobuf/grpc"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"cosmossdk.io/core/appmodule"

	runtimev1 "github.com/cosmos/cosmos-sdk/api/cosmos/base/runtime/v1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/version"
)

type BaseAppOption func(*baseapp.BaseApp)

func (b BaseAppOption) IsAutoGroupType() {}

type appBuilder struct {
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
}

func (a *appBuilder) registerStoreKey(key storetypes.StoreKey) {
	a.storeKeys = append(a.storeKeys, key)
}

func init() {
	appmodule.Register(&runtimev1.Module{},
		appmodule.Provide(
			provideBuilder,
			provideApp,
			provideKVStoreKey,
			provideTransientStoreKey,
			provideMemoryStoreKey,
		),
	)
}

func provideBuilder(moduleBasics map[string]module.AppModuleBasicWiringWrapper) (
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	*appBuilder,
	codec.ProtoCodecMarshaler) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	amino := codec.NewLegacyAmino()

	// build codecs
	for _, wrapper := range moduleBasics {
		wrapper.RegisterInterfaces(interfaceRegistry)
		wrapper.RegisterLegacyAminoCodec(amino)
	}
	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(amino)

	cdc := codec.NewProtoCodec(interfaceRegistry)
	builder := &appBuilder{
		storeKeys:         nil,
		interfaceRegistry: interfaceRegistry,
		cdc:               cdc,
		amino:             amino,
	}

	return interfaceRegistry, cdc, amino, builder, cdc
}

type AppCreator struct {
	app *App
}

func (a *AppCreator) RegisterModules(modules ...module.AppModule) error {
	for _, appModule := range modules {
		if _, ok := a.app.mm.Modules[appModule.Name()]; ok {
			return fmt.Errorf("module named %q already exists", appModule.Name())
		}
		a.app.mm.Modules[appModule.Name()] = appModule
	}
	return nil
}

func (a *AppCreator) Create(logger log.Logger, db dbm.DB, traceStore io.Writer, baseAppOptions ...func(*baseapp.BaseApp)) *App {
	for _, option := range a.app.baseAppOptions {
		baseAppOptions = append(baseAppOptions, option)
	}
	bApp := baseapp.NewBaseApp(a.app.config.AppName, logger, db, baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(a.app.builder.interfaceRegistry)
	bApp.MountStores(a.app.builder.storeKeys...)
	bApp.SetTxHandler(a.app.txHandler)

	a.app.BaseApp = bApp
	return a.app
}

func (a *AppCreator) Finish(loadLatest bool) error {
	if a.app == nil {
		return fmt.Errorf("app not created yet, can't finish")
	}

	configurator := module.NewConfigurator(a.app.builder.cdc, a.app.msgServiceRegistrar, a.app.GRPCQueryRouter())
	a.app.mm.RegisterServices(configurator)
	a.app.mm.SetOrderInitGenesis(a.app.config.InitGenesis...)
	a.app.mm.SetOrderBeginBlockers(a.app.config.BeginBlockers...)
	a.app.mm.SetOrderEndBlockers(a.app.config.EndBlockers...)
	a.app.SetBeginBlocker(a.app.mm.BeginBlock)
	a.app.SetEndBlocker(a.app.mm.EndBlock)
	a.app.SetInitChainer(a.app.InitChainer)

	if loadLatest {
		if err := a.app.LoadLatestVersion(); err != nil {
			return err
		}
	}

	return nil
}

func provideApp(
	config *runtimev1.Module,
	builder *appBuilder,
	modules map[string]module.AppModuleWiringWrapper,
	baseAppOptions []BaseAppOption,
	txHandler tx.Handler,
	msgServiceRegistrar grpc.Server,
) *AppCreator {
	mm := &module.Manager{Modules: map[string]module.AppModule{}}
	for name, wrapper := range modules {
		mm.Modules[name] = wrapper.AppModule
	}
	return &AppCreator{
		app: &App{
			BaseApp:             nil,
			baseAppOptions:      baseAppOptions,
			config:              config,
			builder:             builder,
			mm:                  mm,
			beginBlockers:       nil,
			endBlockers:         nil,
			txHandler:           txHandler,
			msgServiceRegistrar: msgServiceRegistrar,
		},
	}
}

func provideKVStoreKey(key container.ModuleKey, builder *appBuilder) *storetypes.KVStoreKey {
	storeKey := storetypes.NewKVStoreKey(key.Name())
	builder.registerStoreKey(storeKey)
	return storeKey
}

func provideTransientStoreKey(key container.ModuleKey, builder *appBuilder) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	builder.registerStoreKey(storeKey)
	return storeKey
}

func provideMemoryStoreKey(key container.ModuleKey, builder *appBuilder) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	builder.registerStoreKey(storeKey)
	return storeKey
}
