package blockexec_test

import (
	"testing"

	dbm "github.com/cosmos/cosmos-db"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log/v2"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/baseapp/blockexec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func denom(storetypes.MultiStore) string { return sdk.DefaultBondDenom }

func installedParallel(bApp *baseapp.BaseApp) (parallel bool) {
	defer func() { parallel = recover() != nil }()
	bApp.SetDisableBlockGasMeter(false)
	return false
}

func newApp(t *testing.T) *baseapp.BaseApp {
	t.Helper()
	return baseapp.NewBaseApp("blockexec-test", log.NewNopLogger(), dbm.NewMemDB(), nil)
}

// TestApplyDefaultExecutor covers the programmatic wiring path (no flag/app.toml
// bound in appOpts), where WithDefaultExecutor selects the executor.
func TestApplyDefaultExecutor(t *testing.T) {
	t.Run("default falls back to sequential", func(t *testing.T) {
		bApp := newApp(t)
		blockexec.Apply(bApp, simtestutil.AppOptionsMap{}, nil, nil, denom)
		require.False(t, installedParallel(bApp))
	})

	t.Run("WithDefaultExecutor selects block-stm", func(t *testing.T) {
		bApp := newApp(t)
		blockexec.Apply(bApp, simtestutil.AppOptionsMap{}, nil, nil, denom,
			blockexec.WithDefaultExecutor(config.BlockExecutorBlockSTM),
			blockexec.WithDefaultPreEstimate(true),
		)
		require.True(t, installedParallel(bApp))
	})

	t.Run("appOpts flag overrides the default", func(t *testing.T) {
		bApp := newApp(t)
		blockexec.Apply(bApp, simtestutil.AppOptionsMap{
			server.FlagBlockExecutor: config.BlockExecutorSequential,
		}, nil, nil, denom,
			blockexec.WithDefaultExecutor(config.BlockExecutorBlockSTM),
		)
		require.False(t, installedParallel(bApp))
	})
}
