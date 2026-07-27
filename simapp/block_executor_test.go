package simapp

import (
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func TestNewSimApp_BlockExecutorWiring(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		opts := simtestutil.AppOptionsMap{
			flags.FlagHome:           t.TempDir(),
			server.FlagBlockExecutor: "uknown-executor",
		}
		require.PanicsWithError(t, "unknown block executor: uknown-executor", func() {
			NewSimApp(log.NewNopLogger(), dbm.NewMemDB(), true, opts)
		})
	})

	for _, executor := range []string{"", serverconfig.BlockExecutorSequential, serverconfig.BlockExecutorBlockSTM} {
		name := executor
		if name == "" {
			name = "default"
		}
		t.Run(name, func(t *testing.T) {
			opts := simtestutil.AppOptionsMap{flags.FlagHome: t.TempDir()}
			if executor != "" {
				opts[server.FlagBlockExecutor] = executor
			}

			app := NewSimappWithCustomOptions(t, false, SetupOptions{
				Logger:  log.NewNopLogger(),
				DB:      dbm.NewMemDB(),
				AppOpts: opts,
			})

			_, err := app.FinalizeBlock(&abci.RequestFinalizeBlock{
				Height: app.LastBlockHeight() + 1,
				Hash:   app.LastCommitID().Hash,
			})
			require.NoError(t, err)
			_, err = app.Commit()
			require.NoError(t, err)

			if executor == serverconfig.BlockExecutorBlockSTM {
				require.PanicsWithValue(t,
					"Cannot enable block gas meter while parallel execution is configured",
					func() { app.SetDisableBlockGasMeter(false) })
			} else {
				require.NotPanics(t, func() { app.SetDisableBlockGasMeter(false) })
			}
		})
	}
}
