package runtime

import (
	"fmt"

	"github.com/gogo/protobuf/grpc"

	runtimev1alpha1 "github.com/cosmos/cosmos-sdk/api/cosmos/app/runtime/v1alpha1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// BaseAppOption is a container.AutoGroupType which can be used to pass
// BaseApp options into the container. It should be used carefully.
type BaseAppOption func(*baseapp.BaseApp)

// IsAutoGroupType indicates that this is a container.AutoGroupType.
func (b BaseAppOption) IsAutoGroupType() {}

type privateState struct {
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
	basicManager      module.BasicManager
}

func (a *privateState) registerStoreKey(key storetypes.StoreKey) {
	a.storeKeys = append(a.storeKeys, key)
}

func init() {
	// TODO:
	//appmodule.Register(&runtimev1alpha1.Module{},
	//	appmodule.Provide(
	//		provideBuilder,
	//		provideApp,
	//		provideKVStoreKey,
	//		provideTransientStoreKey,
	//		provideMemoryStoreKey,
	//	),
	//)
}

func provideBuilder(moduleBasics map[string]AppModuleBasicWrapper) (
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	*privateState,
	codec.ProtoCodecMarshaler) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	amino := codec.NewLegacyAmino()

	// build codecs
	basicManager := module.BasicManager{}
	for name, wrapper := range moduleBasics {
		basicManager[name] = wrapper
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
		basicManager:      basicManager,
	}

	return interfaceRegistry, cdc, amino, builder, cdc
}

func provideApp(
	config *runtimev1alpha1.Module,
	builder *privateState,
	modules map[string]AppModuleWrapper,
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
			ModuleManager:       mm,
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
