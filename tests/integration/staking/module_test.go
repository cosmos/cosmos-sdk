package staking

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	authKeeper "cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/staking/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authKeeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		), &accountKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.BondedPoolName))
	require.NotNil(t, acc)

	acc = accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.NotBondedPoolName))
	require.NotNil(t, acc)
}
