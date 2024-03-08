package runtime

import (
	"context"
	"encoding/json"

	"golang.org/x/exp/slices"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	coreappmanager "cosmossdk.io/server/v2/core/appmanager"
	corestore "cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

var _ AppI[transaction.Tx] = (*App)(nil)

// AppI is an interface that defines the methods required by the App.
type AppI[T transaction.Tx] interface {
	DeliverBlock(ctx context.Context, block *coreappmanager.BlockRequest[T]) (*coreappmanager.BlockResponse, corestore.WriterMap, error)
	ValidateTx(ctx context.Context, tx T) (coreappmanager.TxResult, error)
	Simulate(ctx context.Context, tx T) (coreappmanager.TxResult, corestore.WriterMap, error)
	Query(ctx context.Context, version uint64, request transaction.Type) (transaction.Type, error)
	QueryWithState(ctx context.Context, state corestore.ReaderMap, request transaction.Type) (transaction.Type, error)

	Logger() log.Logger
	ModuleManager() *MM
	Close() error
}

// App is a wrapper around AppManager and ModuleManager that can be used in hybrid
// app.go/app config scenarios or directly as a servertypes.Application instance.
// To get an instance of *App, *AppBuilder must be requested as a dependency
// in a container which declares the runtime module and the AppBuilder.Build()
// method must be called.
//
// App can be used to create a hybrid app.go setup where some configuration is
// done declaratively with an app config and the rest of it is done the old way.
// See simapp/app_.go for an example of this setup.
type App struct {
	*appmanager.AppManager[transaction.Tx]

	// app manager dependencies
	stf                *stf.STF[transaction.Tx]
	msgRouterBuilder   *stf.MsgRouterBuilder
	queryRouterBuilder *stf.MsgRouterBuilder
	db                 Store // TODO: double check

	// app configuration
	logger    log.Logger
	config    *runtimev2.Module
	appConfig *appv1alpha1.Config

	// modules configuration
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
	moduleManager     *MM
}

// Logger returns the app logger.
func (a *App) Logger() log.Logger {
	return a.logger
}

// ModuleManager returns the module manager.
func (a *App) ModuleManager() *MM {
	return a.moduleManager
}

// DefaultGenesis returns a default genesis from the registered modules.
func (a *App) DefaultGenesis() map[string]json.RawMessage {
	return a.moduleManager.DefaultGenesis(a.cdc)
}

// LoadLatest loads the latest version.
func (a *App) LoadLatest() error {
	return a.db.LoadLatestVersion()
}

// LoadHeight loads a particular height
func (a *App) LoadHeight(height uint64) error {
	return a.db.LoadVersion(height)
}

// Close is called in start cmd to gracefully cleanup resources.
func (a *App) Close() error {
	return nil
}

// RegisterStores registers the provided store keys.
// This method should only be used for registering extra stores
// wiich is necessary for modules that not registered using the app config.
// To be used in combination of RegisterModules.
func (a *App) RegisterStores(keys ...storetypes.StoreKey) error {
	a.storeKeys = append(a.storeKeys, keys...)
	// a.MountStores(keys...)

	return nil
}

// GetStoreKeys returns all the stored store keys.
func (a *App) GetStoreKeys() []storetypes.StoreKey {
	return a.storeKeys
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
// NOTE: This should only be used in testing.
func (a *App) UnsafeFindStoreKey(storeKey string) storetypes.StoreKey {
	i := slices.IndexFunc(a.storeKeys, func(s storetypes.StoreKey) bool { return s.Name() == storeKey })
	if i == -1 {
		return nil
	}

	return a.storeKeys[i]
}

func (a *App) GetStore() Store {
	return a.db
}

func (a *App) GetLogger() log.Logger {
	return a.logger
}
