package runtime

import (
	"encoding/json"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appmodulev2 "cosmossdk.io/core/appmodule/v2"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
)

// App is a wrapper around AppManager and ModuleManager that can be used in hybrid
// app.go/app config scenarios or directly as a servertypes.Application instance.
// To get an instance of *App, *AppBuilder must be requested as a dependency
// in a container which declares the runtime module and the AppBuilder.Build()
// method must be called.
//
// App can be used to create a hybrid app.go setup where some configuration is
// done declaratively with an app config and the rest of it is done the old way.
// See simapp/app_v2.go for an example of this setup.
type App[T transaction.Tx] struct {
	appmanager.AppManager[T]

	// app configuration
	logger log.Logger
	config *runtimev2.Module

	// state
	stf                *stf.STF[T]
	msgRouterBuilder   *stf.MsgRouterBuilder
	queryRouterBuilder *stf.MsgRouterBuilder
	db                 Store
	storeLoader        StoreLoader

	// modules
	interfaceRegistrar registry.InterfaceRegistrar
	amino              registry.AminoRegistrar
	moduleManager      *MM[T]
	queryHandlers      map[string]appmodulev2.Handler // queryHandlers defines the query handlers
}

// Name returns the app name.
func (a *App[T]) Name() string {
	return a.config.AppName
}

// Logger returns the app logger.
func (a *App[T]) Logger() log.Logger {
	return a.logger
}

// ModuleManager returns the module manager.
func (a *App[T]) ModuleManager() *MM[T] {
	return a.moduleManager
}

// DefaultGenesis returns a default genesis from the registered modules.
func (a *App[T]) DefaultGenesis() map[string]json.RawMessage {
	return a.moduleManager.DefaultGenesis()
}

// SetStoreLoader sets the store loader.
func (a *App[T]) SetStoreLoader(loader StoreLoader) {
	a.storeLoader = loader
}

// LoadLatest loads the latest version.
func (a *App[T]) LoadLatest() error {
	return a.storeLoader(a.db)
}

// LoadHeight loads a particular height
func (a *App[T]) LoadHeight(height uint64) error {
	return a.db.LoadVersion(height)
}

// LoadLatestHeight loads the latest height.
func (a *App[T]) LoadLatestHeight() (uint64, error) {
	return a.db.GetLatestVersion()
}

// GetQueryHandlers returns the query handlers.
func (a *App[T]) QueryHandlers() map[string]appmodulev2.Handler {
	return a.queryHandlers
}

// SchemaDecoderResolver returns the module schema resolver.
func (a *App[T]) SchemaDecoderResolver() decoding.DecoderResolver {
	moduleSet := map[string]any{}
	for moduleName, module := range a.moduleManager.Modules() {
		moduleSet[moduleName] = module
	}
	return decoding.ModuleSetDecoderResolver(moduleSet)
}

// Close is called in start cmd to gracefully cleanup resources.
func (a *App[T]) Close() error {
	return nil
}
