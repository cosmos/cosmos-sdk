package baseapp

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"

	"github.com/cosmos/cosmos-sdk/baseapp/config"
	"github.com/cosmos/cosmos-sdk/baseapp/state"
)

// Ensures that error checks are performed before sealing the app.
// Please see https://github.com/cosmos/cosmos-sdk/issues/18726
func TestNilCmsCheckBeforeSeal(t *testing.T) {
	app := new(BaseApp)
	app.stateManager = state.NewManager(config.GasConfig{})

	// 1. Invoking app.Init with a nil cms MUST not seal the app
	// and should return an error firstly, which can later be reversed.
	for range 10 { // N times, the app shouldn't be sealed.
		err := app.Init()
		require.Error(t, err)
		require.Contains(t, err.Error(), "commit multi-store must not be nil")
		require.False(t, app.IsSealed(), "the app MUST not be sealed")
	}

	// 2. Now that we've figured out and gotten back an error, let's rectify the problem.
	// and we should be able to set the commit multistore then reinvoke app.Init successfully!
	db := dbm.NewMemDB()
	logger := log.NewTestLogger(t)
	app.cms = store.NewCommitMultiStore(db, logger, storemetrics.NewNoOpMetrics())
	err := app.Init()
	require.Nil(t, err, "app.Init MUST now succeed")
	require.True(t, app.IsSealed(), "the app must now be sealed")

	// 3. Now we should expect a panic because the app is sealed.
	require.Panics(t, func() {
		_ = app.Init()
	})
}
