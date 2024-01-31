package baseapp

import (
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
