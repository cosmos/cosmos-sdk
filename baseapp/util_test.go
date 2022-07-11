package baseapp

import (
	runtimev1alpha1 "cosmossdk.io/api/cosmos/app/runtime/v1alpha1"
	appv1alpha1 "cosmossdk.io/api/cosmos/app/v1alpha1"
	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	bankmodulev1 "cosmossdk.io/api/cosmos/bank/module/v1"
	"cosmossdk.io/core/appconfig"
	"cosmossdk.io/depinject"

	"github.com/cosmos/cosmos-sdk/snapshots"
	"github.com/cosmos/cosmos-sdk/types"
)

// TODO: Can be removed once we move all middleware tests into x/auth/middleware
// ref: #https://github.com/cosmos/cosmos-sdk/issues/10282

// CheckState is an exported method to be able to access baseapp's
// checkState in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) CheckState() *state {
	return app.checkState
}

// DeliverState is an exported method to be able to access baseapp's
// deliverState in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) DeliverState() *state {
	return app.deliverState
}

// CMS is an exported method to be able to access baseapp's cms in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) CMS() types.CommitMultiStore {
	return app.cms
}

// GetSnapshotManager() is an exported method to be able to access baseapp's snapshot
// manager in tests.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) GetSnapshotManager() *snapshots.Manager {
	return app.snapshotManager
}

// GetMaximumBlockGas return maximum blocks gas.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) GetMaximumBlockGas(ctx types.Context) uint64 {
	return app.getMaximumBlockGas(ctx)
}

// GetName return name.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) GetName() string {
	return app.name
}

// CreateQueryContext calls app's createQueryContext.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) CreateQueryContext(height int64, prove bool) (types.Context, error) {
	return app.createQueryContext(height, prove)
}

// MinGasPrices returns minGasPrices.
//
// This method is only accessible in baseapp tests.
func (app *BaseApp) MinGasPrices() types.DecCoins {
	return app.minGasPrices
}

func MakeTestConfig() depinject.Config {
	return appconfig.Compose(&appv1alpha1.Config{
		Modules: []*appv1alpha1.ModuleConfig{
			{
				Name: "runtime",
				Config: appconfig.WrapAny(&runtimev1alpha1.Module{
					AppName: "BaseAppApp",
					BeginBlockers: []string{
						"auth",
						"bank",
					},
					EndBlockers: []string{
						"auth",
						"bank",
					},
					OverrideStoreKeys: []*runtimev1alpha1.StoreKeyConfig{
						{
							ModuleName: "auth",
							KvStoreKey: "acc",
						},
					},
					InitGenesis: []string{
						"auth",
						"bank",
					},
				}),
			},
			{
				Name: "auth",
				Config: appconfig.WrapAny(&authmodulev1.Module{
					Bech32Prefix: "cosmos",
					ModuleAccountPermissions: []*authmodulev1.ModuleAccountPermission{
						{Account: "fee_collector"},
					},
				}),
			},
			{
				Name:   "bank",
				Config: appconfig.WrapAny(&bankmodulev1.Module{}),
			},
		},
	})
}
