package runtime

import (
	"golang.org/x/exp/slices"

	runtimev2 "cosmossdk.io/api/cosmos/app/runtime/v2"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	"cosmossdk.io/server/v2/appmanager"
	"cosmossdk.io/server/v2/stf"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	stf                *stf.STF[transaction.Tx]
	msgRouterBuilder   *stf.MsgRouterBuilder
	queryRouterBuilder *stf.MsgRouterBuilder
	db                 Store

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

	appManagerBuilder := appmanager.Builder[transaction.Tx]{
		STF:                a.stf,
		DB:                 a.db,
		ValidateTxGasLimit: a.config.GasConfig.ValidateTxGasLimit,
		QueryGasLimit:      a.config.GasConfig.QueryGasLimit,
		SimulationGasLimit: a.config.GasConfig.SimulationGasLimit,
	}

	appManager, err := appManagerBuilder.Build()
	if err != nil {
		return err
	}
	a.AppManager = appManager

	return nil
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
