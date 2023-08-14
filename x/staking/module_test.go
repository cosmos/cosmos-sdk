package staking_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authKeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authKeeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			testutil.AppConfig,
			depinject.Supply(log.NewNopLogger()),
		), &accountKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.BondedPoolName))
	require.NotNil(t, acc)

	acc = accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.NotBondedPoolName))
	require.NotNil(t, acc)
}
