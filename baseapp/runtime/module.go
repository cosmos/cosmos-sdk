package runtime

import (
	"fmt"

	"github.com/gogo/protobuf/grpc"

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
)

type BaseAppOption func(*baseapp.BaseApp)

func (b BaseAppOption) IsAutoGroupType() {}

type privateState struct {
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
	moduleBasics      map[string]module.AppModuleBasicWiringWrapper
}

func (a *privateState) registerStoreKey(key storetypes.StoreKey) {
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
	*privateState,
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
	builder := &privateState{
		storeKeys:         nil,
		interfaceRegistry: interfaceRegistry,
		cdc:               cdc,
		amino:             amino,
		moduleBasics:      moduleBasics,
	}

	return interfaceRegistry, cdc, amino, builder, cdc
}

func provideApp(
	config *runtimev1.Module,
	builder *privateState,
	modules map[string]module.AppModuleWiringWrapper,
	baseAppOptions []BaseAppOption,
	txHandler tx.Handler,
	msgServiceRegistrar grpc.Server,
) *AppBuilder {
	mm := &module.Manager{Modules: map[string]module.AppModule{}}
	for name, wrapper := range modules {
		mm.Modules[name] = wrapper.AppModule
	}
	return &AppBuilder{
		app: &App{
			BaseApp:             nil,
			baseAppOptions:      baseAppOptions,
			config:              config,
			privateState:        builder,
			mm:                  mm,
			beginBlockers:       nil,
			endBlockers:         nil,
			txHandler:           txHandler,
			msgServiceRegistrar: msgServiceRegistrar,
		},
	}
}

func provideKVStoreKey(key container.ModuleKey, builder *privateState) *storetypes.KVStoreKey {
	storeKey := storetypes.NewKVStoreKey(key.Name())
	builder.registerStoreKey(storeKey)
	return storeKey
}

func provideTransientStoreKey(key container.ModuleKey, builder *privateState) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	builder.registerStoreKey(storeKey)
	return storeKey
}

func provideMemoryStoreKey(key container.ModuleKey, builder *privateState) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	builder.registerStoreKey(storeKey)
	return storeKey
}
