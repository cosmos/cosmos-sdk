package runtime

import (
	"fmt"

	"github.com/gogo/protobuf/grpc"

	"github.com/cosmos/cosmos-sdk/container"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// BaseAppOption is a container.AutoGroupType which can be used to pass
// BaseApp options into the container. It should be used carefully.
type BaseAppOption func(*baseapp.BaseApp)

// IsManyPerContainerType indicates that this is a container.ManyPerContainerType.
func (b BaseAppOption) IsManyPerContainerType() {}

// appWrapper is used to pass around an instance of *App internally between
// runtime dependency inject providers that is partially constructed (no
// baseapp yet).
type appWrapper *App

func init() {
	appmodule.Register(&runtimev1alpha1.Module{},
		appmodule.Provide(
			provideCodecs,
			provideAppBuilder,
			provideKVStoreKey,
			provideTransientStoreKey,
			provideMemoryStoreKey,
		),
	)
}

func provideCodecs(moduleBasics map[string]AppModuleBasicWrapper) (
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	appWrapper,
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
	app := &App{
		storeKeys:         nil,
		interfaceRegistry: interfaceRegistry,
		cdc:               cdc,
		amino:             amino,
		basicManager:      basicManager,
	}

	return interfaceRegistry, cdc, amino, app, cdc
}

type appInputs struct {
	container.In

	Config              *runtimev1alpha1.Module
	App                 appWrapper
	Modules             map[string]AppModuleWrapper
	BaseAppOptions      []BaseAppOption
	MsgServiceRegistrar grpc.Server `optional:"true"`
}

func provideAppBuilder(inputs appInputs) *AppBuilder {
	mm := &module.Manager{Modules: map[string]module.AppModule{}}
	for name, wrapper := range inputs.Modules {
		mm.Modules[name] = wrapper.AppModule
	}
	app := inputs.App
	app.baseAppOptions = inputs.BaseAppOptions
	app.config = inputs.Config
	app.ModuleManager = mm
	app.msgServiceRegistrar = inputs.MsgServiceRegistrar
	return &AppBuilder{app: app}
}

func registerStoreKey(wrapper appWrapper, key storetypes.StoreKey) {
	wrapper.storeKeys = append(wrapper.storeKeys, key)
}

func provideKVStoreKey(key container.ModuleKey, app appWrapper) *storetypes.KVStoreKey {
	storeKey := storetypes.NewKVStoreKey(key.Name())
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideTransientStoreKey(key container.ModuleKey, app appWrapper) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideMemoryStoreKey(key container.ModuleKey, app appWrapper) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}
