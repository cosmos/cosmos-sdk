package gov_test

import (
	"testing"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	var accountKeeper authkeeper.AccountKeeper
	app, err := simtestutil.SetupWithConfiguration(
		configurator.NewAppConfig(
			configurator.ParamsModule(),
			configurator.AuthModule(),
			configurator.StakingModule(),
			configurator.BankModule(),
			configurator.GovModule(),
			configurator.ConsensusModule(),
		),
		simtestutil.DefaultStartUpConfig(),
		&accountKeeper,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	acc := accountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
