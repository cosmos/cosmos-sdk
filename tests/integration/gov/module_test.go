package gov_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	_ "github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	_ "github.com/cosmos/cosmos-sdk/x/mint"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper
	app, err := simtestutil.SetupAtGenesisWithNopLogger(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.StakingModule(),
			configurator.BankModule(),
			configurator.GovModule(),
			configurator.DistributionModule(),
			configurator.ConsensusModule(),
		),
		&accountKeeper,
	)
	assert.NilError(t, err)

	ctx := app.NewContext(false)
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
