package runtime

import (
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
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

func init() {
	appmodule.Register(&runtimev1alpha1.Module{},
		appmodule.Provide(
			ProvideCodecs,
			ProvideKVStoreKey,
			ProvideTransientStoreKey,
			ProvideMemoryStoreKey,
			ProvideDeliverTx,
		),
		appmodule.Invoke(SetupAppBuilder),
	)
}

func ProvideCodecs(moduleBasics map[string]AppModuleBasicWrapper) (
	codectypes.InterfaceRegistry,
	codec.Codec,
	*codec.LegacyAmino,
	*AppBuilder,
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
	app := &AppBuilder{
		&App{
			storeKeys:         nil,
			interfaceRegistry: interfaceRegistry,
			cdc:               cdc,
			amino:             amino,
			basicManager:      basicManager,
			msgServiceRouter:  msgServiceRouter,
		},
	}

	return interfaceRegistry, cdc, amino, app, cdc, msgServiceRouter
}

type appInputs struct {
	depinject.In

	AppConfig      *appv1alpha1.Config
	Config         *runtimev1alpha1.Module
	AppBuilder     *AppBuilder
	Modules        map[string]AppModuleWrapper
	BaseAppOptions []BaseAppOption
	AppModules     map[string]appmodule.AppModule
}

func SetupAppBuilder(inputs appInputs) {
	mm := &module.Manager{Modules: map[string]module.AppModule{}}
	for name, wrapper := range inputs.Modules {
		mm.Modules[name] = wrapper.AppModule
	}
	app := inputs.AppBuilder.app
	app.baseAppOptions = inputs.BaseAppOptions
	app.config = inputs.Config
	app.ModuleManager = mm
	app.appConfig = inputs.AppConfig
	app.appModules = inputs.AppModules
}

func registerStoreKey(wrapper *AppBuilder, key storetypes.StoreKey) {
	wrapper.app.storeKeys = append(wrapper.app.storeKeys, key)
}

func storeKeyOverride(config *runtimev1alpha1.Module, moduleName string) *runtimev1alpha1.StoreKeyConfig {
	for _, cfg := range config.OverrideStoreKeys {
		if cfg.ModuleName == moduleName {
			return cfg
		}
	}
	return nil
}

func ProvideKVStoreKey(config *runtimev1alpha1.Module, key depinject.ModuleKey, app *AppBuilder) *storetypes.KVStoreKey {
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

func ProvideTransientStoreKey(key depinject.ModuleKey, app *AppBuilder) *storetypes.TransientStoreKey {
	storeKey := storetypes.NewTransientStoreKey(fmt.Sprintf("transient:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func ProvideMemoryStoreKey(key depinject.ModuleKey, app *AppBuilder) *storetypes.MemoryStoreKey {
	storeKey := storetypes.NewMemoryStoreKey(fmt.Sprintf("memory:%s", key.Name()))
	registerStoreKey(app, storeKey)
	return storeKey
}

func ProvideDeliverTx(appBuilder *AppBuilder) func(abci.RequestDeliverTx) abci.ResponseDeliverTx {
	return func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
		return appBuilder.app.BaseApp.DeliverTx(tx)
	}
}
