package runtime

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/depinject"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// BaseAppOption is a depinject.AutoGroupType which can be used to pass
// BaseApp options into the depinject. It should be used carefully.
type BaseAppOption func(*baseapp.BaseApp)

// IsManyPerContainerType indicates that this is a depinject.ManyPerContainerType.
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
			provideDeliverTx,
		),
	)
}

func provideCodecs(moduleBasics map[string]AppModuleBasicWrapper) (
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	appWrapper,
	codec.ProtoCodecMarshaler,
	*baseapp.MsgServiceRouter,
) {
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
	msgServiceRouter := baseapp.NewMsgServiceRouter()
	app := &App{
		storeKeys:         nil,
		interfaceRegistry: interfaceRegistry,
		cdc:               cdc,
		amino:             amino,
		basicManager:      basicManager,
		msgServiceRouter:  msgServiceRouter,
	}

	return interfaceRegistry, cdc, amino, app, cdc, msgServiceRouter
}

type appInputs struct {
	depinject.In

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

func storeKeyOverride(config *runtimev1alpha1.Module, moduleName string) *runtimev1alpha1.StoreKeyConfig {
	for _, cfg := range config.OverrideStoreKeys {
		if cfg.ModuleName == moduleName {
			return cfg
		}
	}
	return nil
}

func provideKVStoreKey(config *runtimev1alpha1.Module, key depinject.ModuleKey, app appWrapper) *storetypes.KVStoreKey {
	override := storeKeyOverride(config, key.Name())

	var storeKeyName string
	if override != nil {
		storeKeyName = override.KvStoreKey
	} else {
		storeKeyName = key.Name()
	}

	storeKey := storetypes.NewKVStoreKey(storeKeyName)
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideTransientStoreKey(key depinject.ModuleKey, app appWrapper) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideMemoryStoreKey(key depinject.ModuleKey, app appWrapper) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func provideDeliverTx(app appWrapper) func(abci.RequestDeliverTx) abci.ResponseDeliverTx {
	return func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
		return app.BaseApp.DeliverTx(tx)
	}
}
