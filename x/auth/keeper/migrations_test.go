package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/cosmos/cosmos-sdk/x/auth/keeper"

	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types3 "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateVestingAccounts(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(310))
	delegatorAddr := addrs[0]

	_, valAddr := createValidator(t, ctx, app, 300)
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)

	baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
	vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))

	delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, time.Now().Unix())

	_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
	require.NoError(t, err)

	// We introduce the bug
	delayedAccount.DelegatedFree = nil
	delayedAccount.DelegatedVesting = nil
	app.AccountKeeper.SetAccount(ctx, delayedAccount)

	migrator := keeper.NewMigrator(app.AccountKeeper)
	err = migrator.Migrate1to2(ctx)
	require.NoError(t, err)

	savedDelayedAccount := app.AccountKeeper.GetAccount(ctx, delegatorAddr)
	vDA, ok := savedDelayedAccount.(exported.VestingAccount)
	require.True(t, ok)
	require.Equal(t, vestedCoins, vDA.GetDelegatedVesting())
}

func createValidator(t *testing.T, ctx sdk.Context, app *simapp.SimApp, powers int64) (sdk.AccAddress, sdk.ValAddress) {
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(30000000))
	valAddrs := simapp.ConvertAddrsToValAddrs(addrs)
	pks := simapp.CreateTestPubKeys(1)
	cdc := simapp.MakeTestEncodingConfig().Marshaler

	app.StakingKeeper = stakingkeeper.NewKeeper(
		cdc,
		app.GetKey(stakingtypes.StoreKey),
		app.AccountKeeper,
		app.BankKeeper,
		app.GetSubspace(stakingtypes.ModuleName),
	)

	val1, err := stakingtypes.NewValidator(valAddrs[0], pks[0], stakingtypes.Description{})
	require.NoError(t, err)

	app.StakingKeeper.SetValidator(ctx, val1)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, val1)
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)

	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers), stakingtypes.Unbonded, val1, true)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return addrs[0], valAddrs[0]
}
