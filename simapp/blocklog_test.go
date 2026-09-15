package simapp

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/server/blocklog"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

// TestBlockLogCapturesInjectedKeeperLogger runs real blocks through SimApp
// with the root logger wrapped for capture and checks that x/bank, whose
// keeper stores the logger it received at construction, has its per-block
// debug line attributed to the right height.
func TestBlockLogCapturesInjectedKeeperLogger(t *testing.T) {
	store, err := blocklog.Open(t.TempDir(), 10, log.NewNopLogger())
	require.NoError(t, err)

	app := NewSimappWithCustomOptions(t, false, SetupOptions{
		Logger:  blocklog.Wrap(log.NewNopLogger(), store, log.ModuleKey, "server"),
		DB:      dbm.NewMemDB(),
		AppOpts: simtestutil.NewAppOptionsWithFlagHome(t.TempDir()),
	})

	now := time.Now()
	for h := int64(1); h <= 3; h++ {
		_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: h, Time: now.Add(time.Duration(h) * time.Second)})
		require.NoError(t, err)
		_, err = app.Commit()
		require.NoError(t, err)
	}
	require.Equal(t, int64(3), store.LastClosed())

	for h := int64(2); h <= 3; h++ {
		res, err := store.Query(blocklog.Request{Start: h, End: h, Level: "debug", Module: "x/bank"})
		require.NoError(t, err)
		require.NotEmpty(t, res.Entries, "height %d", h)
		for _, e := range res.Entries {
			require.Equal(t, h, e.Height)
			require.Equal(t, "x/bank", e.Module)
		}
		require.Equal(t, "minted coins from module account", res.Entries[0].Msg)
	}
}
