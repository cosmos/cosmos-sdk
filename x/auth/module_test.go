package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper keeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesisWithNopLogger(testutil.AppConfig, &accountKeeper)
	require.NoError(t, err)

	ctx := app.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, types.NewModuleAddress(types.FeeCollectorName))
	require.NotNil(t, acc)
}
