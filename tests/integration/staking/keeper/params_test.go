package keeper

import (
	"testing"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"gotest.tools/v3/assert"

	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestParams(t *testing.T) {
	var stakingKeeper *keeper.Keeper
	app, err := simtestutil.SetupWithConfiguration(
		configurator.NewAppConfig(
			configurator.BankModule(),
			configurator.TxModule(),
			configurator.StakingModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.AuthModule(),
		),
		simtestutil.DefaultStartUpConfig(),
		&stakingKeeper)
	assert.NilError(t, err)
	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})

	expParams := types.DefaultParams()

	// check that the empty keeper loads the default
	resParams := stakingKeeper.GetParams(ctx)
	assert.Assert(t, expParams.Equal(resParams))

	// modify a params, save, and retrieve
	expParams.MaxValidators = 777
	stakingKeeper.SetParams(ctx, expParams)
	resParams = stakingKeeper.GetParams(ctx)
	assert.Assert(t, expParams.Equal(resParams))
}
