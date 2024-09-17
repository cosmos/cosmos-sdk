package runtime

import (
	"encoding/json"
	"errors"
	"slices"

	gogoproto "github.com/cosmos/gogoproto/proto"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	"cosmossdk.io/core/registry"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
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
	*appmanager.AppManager[T]

	// app manager dependencies
	stf                *stf.STF[T]
	msgRouterBuilder   *stf.MsgRouterBuilder
	queryRouterBuilder *stf.MsgRouterBuilder
	db                 Store

	// app configuration
	logger log.Logger
	config *runtimev2.Module

	// modules configuration
	storeKeys          []string
	interfaceRegistrar registry.InterfaceRegistrar
	amino              registry.AminoRegistrar
	moduleManager      *MM[T]

	// GRPCMethodsToMessageMap maps gRPC method name to a function that decodes the request
	// bytes into a gogoproto.Message, which then can be passed to appmanager.
	GRPCMethodsToMessageMap map[string]func() gogoproto.Message

	storeLoader StoreLoader
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

// Close is called in start cmd to gracefully cleanup resources.
func (a *App[T]) Close() error {
	return nil
}

// GetStoreKeys returns all the app store keys.
func (a *App[T]) GetStoreKeys() []string {
	return a.storeKeys
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
// NOTE: This should only be used in testing.
func (a *App[T]) UnsafeFindStoreKey(storeKey string) (string, error) {
	i := slices.IndexFunc(a.storeKeys, func(s string) bool { return s == storeKey })
	if i == -1 {
		return "", errors.New("store key not found")
	}

	return a.storeKeys[i], nil
}

// GetStore returns the app store.
func (a *App[T]) GetStore() Store {
	return a.db
}

func (a *App[T]) GetAppManager() *appmanager.AppManager[T] {
	return a.AppManager
}

func (a *App[T]) GetGPRCMethodsToMessageMap() map[string]func() gogoproto.Message {
	return a.GRPCMethodsToMessageMap
}
