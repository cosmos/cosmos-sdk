package runtime

import (
	"fmt"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"golang.org/x/exp/slices"
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

	Config         *runtimev1alpha1.Module
	App            appWrapper
	Modules        map[string]AppModuleWrapper
	BaseAppOptions []BaseAppOption
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
	return &AppBuilder{app: app}
}

func registerStoreKey(wrapper appWrapper, key storetypes.StoreKey) {
	wrapper.storeKeys = append(wrapper.storeKeys, key)
}

func storeKeyName(config *runtimev1alpha1.Module, moduleKey container.ModuleKey) string {
	i := slices.IndexFunc(config.ModuleStoreKeys, func(msk *runtimev1alpha1.ModuleStoreKey) bool {
		return msk.ModuleName == moduleKey.Name()
	})
	if i == -1 {
		return moduleKey.Name()
	}
	return config.ModuleStoreKeys[i].StoreKey
}

func provideKVStoreKey(config *runtimev1alpha1.Module, key container.ModuleKey, app appWrapper) *storetypes.KVStoreKey {
	storeKey := storetypes.NewKVStoreKey(storeKeyName(config, key))
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideTransientStoreKey(config *runtimev1alpha1.Module, key container.ModuleKey, app appWrapper) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", storeKeyName(config, key)))
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideMemoryStoreKey(config *runtimev1alpha1.Module, key container.ModuleKey, app appWrapper) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", storeKeyName(config, key)))
	registerStoreKey(app, storeKey)
	return storeKey
}
