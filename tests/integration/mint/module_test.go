package mint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/x/mint/types"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper

	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			AppConfig,
			depinject.Supply(log.NewNopLogger()),
		), &accountKeeper)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
