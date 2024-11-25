package gov_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	_ "cosmossdk.io/x/accounts"
	"cosmossdk.io/x/gov/types"
	_ "cosmossdk.io/x/mint"
	_ "cosmossdk.io/x/protocolpool"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesis(
		depinject.Configs(
			configurator.NewAppConfig(
				configurator.AccountsModule(),
				configurator.AuthModule(),
				configurator.StakingModule(),
				configurator.BankModule(),
				configurator.GovModule(),
				configurator.ConsensusModule(),
				configurator.ProtocolPoolModule(),
			),
			depinject.Supply(log.NewNopLogger()),
		),
		&accountKeeper,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
