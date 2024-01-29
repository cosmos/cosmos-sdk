package runtime

import (
	"golang.org/x/exp/slices"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	coreappmanager "cosmossdk.io/server/v2/core/appmanager"
	"cosmossdk.io/server/v2/core/mempool"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

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
	stf                 *stf.STF[transaction.Tx]
	msgRouterBuilder    *stf.MsgRouterBuilder
	mempool             mempool.Mempool[transaction.Tx]
	prepareBlockHandler coreappmanager.PrepareHandler[transaction.Tx]
	verifyBlockHandler  coreappmanager.ProcessHandler[transaction.Tx]
	store               store.Store

	// app configuration
	logger    log.Logger
	config    *runtimev2.Module
	appConfig *appv1alpha1.Config

	// modules configuration
	configurator      module.Configurator
	storeKeys         []storetypes.StoreKey
	interfaceRegistry codectypes.InterfaceRegistry
	cdc               codec.Codec
	amino             *codec.LegacyAmino
	moduleManager     *MMv2
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

// Load finishes all initialization operations and loads the app.
func (a *App) Load() error {
	// set defaults
	if a.mempool == nil {
		// a.mempool = mempool.NewBasicMempool()
	}

	if a.prepareBlockHandler == nil {
		// a.prepareBlockHandler = appmanager.DefaultPrepareBlockHandler
	}

	if a.verifyBlockHandler == nil {
		// a.verifyBlockHandler = appmanager.DefaultProcessBlockHandler
	}

	appManagerBuilder := appmanager.Builder[transaction.Tx]{
		STF:                 a.stf,
		DB:                  a.store,
		ValidateTxGasLimit:  a.config.GasConfig.ValidateTxGasLimit,
		QueryGasLimit:       a.config.GasConfig.QueryGasLimit,
		SimulationGasLimit:  a.config.GasConfig.SimulationGasLimit,
		PrepareBlockHandler: a.prepareBlockHandler,
		VerifyBlockHandler:  a.verifyBlockHandler,
	}

	appManager, err := appManagerBuilder.Build()
	if err != nil {
		return err
	}

	a.AppManager = appManager

	return nil
}

// Configurator returns the app's configurator.
func (a *App) Configurator() module.Configurator {
	return a.configurator
}

// LoadHeight loads a particular height
func (a *App) LoadHeight(height int64) error {
	return nil // TODO
}

// GetStoreKeys returns all the stored store keys.
func (a *App) GetStoreKeys() []storetypes.StoreKey {
	return a.storeKeys
}

// UnsafeFindStoreKey fetches a registered StoreKey from the App in linear time.
//
// NOTE: This should only be used in testing.
func (a *App) UnsafeFindStoreKey(storeKey string) storetypes.StoreKey {
	i := slices.IndexFunc(a.storeKeys, func(s storetypes.StoreKey) bool { return s.Name() == storeKey })
	if i == -1 {
		return nil
	}

	return a.storeKeys[i]
}

// TODO
// Genesis (NO BASIC MANAGER), Migrations
