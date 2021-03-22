package migrations_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"

	"github.com/cosmos/cosmos-sdk/x/staking"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	types3 "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateVestingAccounts(t *testing.T) {
	testCases := []struct {
		name        string
		prepareFunc func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress)
		expVested   int64
		expFree     int64
	}{
		{
			"delayed vesting has vested, multiple delegations less than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			0,
			200,
		},
		{
			"delayed vesting has vested, single delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			0,
			200,
		},
		{
			"delayed vesting has vested, multiple delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			0,
			200,
		},
		{
			"delayed vesting has not vested, single delegations  which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			200,
			0,
		},
		{
			"delayed vesting has not vested, multiple delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			200,
			0,
		},
		{
			"not end time",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
				_, err = app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(100), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			300,
			0,
		},
		{
			"delayed vesting has not vested, single delegation greater than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			300,
			0,
		},
		{
			"delayed vesting has vested, single delegation greater than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			0,
			300,
		},
		{
			"continuous vesting, start time after blocktime",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				startTime := ctx.BlockTime().AddDate(1, 0, 0).Unix()
				endTime := ctx.BlockTime().AddDate(2, 0, 0).Unix()
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			300,
			0,
		},
		{
			"continuous vesting, start time passed but not ended",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				startTime := ctx.BlockTime().AddDate(-1, 0, 0).Unix()
				endTime := ctx.BlockTime().AddDate(2, 0, 0).Unix()
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			200,
			100,
		},
		{
			"continuous vesting, start time and endtime passed",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				startTime := ctx.BlockTime().AddDate(-2, 0, 0).Unix()
				endTime := ctx.BlockTime().AddDate(-1, 0, 0).Unix()
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			0,
			300,
		},
		{
			"periodic vesting account, yet to be vested, some rewards delegated",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(100)))

				start := ctx.BlockTime().Unix() + int64(time.Hour/time.Second)

				periods := []types.Period{
					{
						Length: int64((24 * time.Hour) / time.Second),
						Amount: vestedCoins,
					},
				}

				account := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, start, periods)

				app.AccountKeeper.SetAccount(ctx, account)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(150), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			100,
			50,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{
				Time: time.Now(),
			})

			addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(310))
			delegatorAddr := addrs[0]

			_, valAddr := createValidator(t, ctx, app, 300)
			validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
			require.True(t, found)

			tc.prepareFunc(app, ctx, validator, delegatorAddr)

			// We introduce the bug
			savedAccount := app.AccountKeeper.GetAccount(ctx, delegatorAddr)
			vestingAccount, ok := savedAccount.(exported.VestingAccount)
			require.True(t, ok)
			require.NoError(t, introduceTrackingBug(ctx, vestingAccount, app))

			migrator := authkeeper.NewMigrator(app.AccountKeeper, app.GRPCQueryRouter())
			require.NoError(t, migrator.Migrate1to2(ctx))

			var expVested sdk.Coins
			var expFree sdk.Coins

			if tc.expVested != 0 {
				expVested = sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(tc.expVested)))
			}

			if tc.expFree != 0 {
				expFree = sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(tc.expFree)))
			}

			trackingCorrected(
				ctx,
				t,
				app.AccountKeeper,
				savedAccount.GetAddress(),
				expVested,
				expFree,
			)
		})
	}

}

func trackingCorrected(ctx sdk.Context, t *testing.T, ak authkeeper.AccountKeeper, addr sdk.AccAddress, expDelVesting sdk.Coins, expDelFree sdk.Coins) {
	baseAccount := ak.GetAccount(ctx, addr)
	vDA, ok := baseAccount.(exported.VestingAccount)
	require.True(t, ok)

	vestedOk := expDelVesting.IsEqual(vDA.GetDelegatedVesting())
	freeOk := expDelFree.IsEqual(vDA.GetDelegatedFree())
	require.True(t, vestedOk, vDA.GetDelegatedVesting().String())
	require.True(t, freeOk, vDA.GetDelegatedFree().String())
}

func introduceTrackingBug(ctx sdk.Context, vesting exported.VestingAccount, app *simapp.SimApp) error {
	switch t := vesting.(type) {
	case *types.DelayedVestingAccount:
		t.DelegatedFree = nil
		t.DelegatedVesting = nil
		app.AccountKeeper.SetAccount(ctx, t)
	case *types.ContinuousVestingAccount:
		t.DelegatedFree = nil
		t.DelegatedVesting = nil
		app.AccountKeeper.SetAccount(ctx, t)
	case *types.PeriodicVestingAccount:
		t.DelegatedFree = nil
		t.DelegatedVesting = nil
		app.AccountKeeper.SetAccount(ctx, t)
	default:
		return fmt.Errorf("expected vesting account, found %t", t)
	}

	return nil
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
