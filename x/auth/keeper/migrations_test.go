package keeper_test

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
		tokenAmount int64
		expVested   int64
		expFree     int64
		blockTime   int64
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
			300,
			0,
			200,
			0,
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
			300,
			0,
			200,
			0,
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
			300,
			0,
			200,
			0,
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
			300,
			200,
			0,
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
			300,
			200,
			0,
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
			300,
			0,
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
			300,
			0,
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
			300,
			0,
			300,
			0,
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
			300,
			0,
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
			300,
			200,
			100,
			0,
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
			300,
			0,
			300,
			0,
		},
		{
			"periodic vesting, start time and endtime passed",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {

				startTime := ctx.BlockTime().AddDate(-1, 0, 0).Unix()
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))

				periods := []types.Period{
					{
						Length: 1,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(100))),
					},
				}

				delayedAccount := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, startTime, periods)

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			300,
			0,
			300,
			0,
		},
		{
			"periodic vesting account, nothing has vested yet",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				startTime := int64(1601042400)
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(3666666670000)))
				periods := []types.Period{
					{
						Length: 31536000,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1833333335000))),
					},
					{
						Length: 15638400,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
					{
						Length: 15897600,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
				}

				delayedAccount := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, startTime, periods)

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				// delegation of the original vesting
				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(3666666670000), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			3666666670000,
			3666666670000,
			0,
			0,
		},
		{
			"periodic vesting account, all has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				startTime := int64(1601042400)
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(3666666670000)))
				periods := []types.Period{
					{
						Length: 31536000,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1833333335000))),
					},
					{
						Length: 15638400,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
					{
						Length: 15897600,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
				}

				delayedAccount := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, startTime, periods)

				ctx = ctx.WithBlockTime(time.Unix(1601042400+31536000+15897600+15897600+1, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				// delegation of the original vesting
				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(3666666670000), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			3666666670000,
			0,
			3666666670000,
			1601042400 + 31536000 + 15897600 + 15897600 + 1,
		},
		{
			"periodic vesting account, first period has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				startTime := int64(1601042400)
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(3666666670000)))
				periods := []types.Period{
					{
						Length: 31536000,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1833333335000))),
					},
					{
						Length: 15638400,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
					{
						Length: 15897600,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
				}

				delayedAccount := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, startTime, periods)

				ctx = ctx.WithBlockTime(time.Unix(1601042400+31536000+1, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				// delegation of the original vesting
				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(3666666670000), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			3666666670000,
			3666666670000 - 1833333335000,
			1833333335000,
			1601042400 + 31536000 + 1,
		},
		{
			"periodic vesting account, first 2 period has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				startTime := int64(1601042400)
				baseAccount := types3.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(3666666670000)))
				periods := []types.Period{
					{
						Length: 31536000,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1833333335000))),
					},
					{
						Length: 15638400,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
					{
						Length: 15897600,
						Amount: sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(916666667500))),
					},
				}

				delayedAccount := types.NewPeriodicVestingAccount(baseAccount, vestedCoins, startTime, periods)

				ctx = ctx.WithBlockTime(time.Unix(1601042400+31536000+15638400+1, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				// delegation of the original vesting
				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(3666666670000), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			3666666670000,
			3666666670000 - 1833333335000 - 916666667500,
			1833333335000 + 916666667500,
			1601042400 + 31536000 + 15638400 + 1,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{
				Time: time.Now(),
			})

			addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(tc.tokenAmount+10))
			delegatorAddr := addrs[0]

			_, valAddr := createValidator(t, ctx, app, tc.tokenAmount*2)
			validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
			require.True(t, found)

			tc.prepareFunc(app, ctx, validator, delegatorAddr)

			if tc.blockTime != 0 {
				ctx = ctx.WithBlockTime(time.Unix(tc.blockTime, 0))
			}

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
	require.True(t, vestedOk, fmt.Sprint("delegated vesting ", vDA.GetDelegatedVesting().String()))
	require.True(t, freeOk, fmt.Sprint("delegated free ", vDA.GetDelegatedFree().String()))
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
	require.NoError(t, app.StakingKeeper.SetValidatorByConsAddr(ctx, val1))
	app.StakingKeeper.SetNewValidatorByPowerIndex(ctx, val1)

	_, _ = app.StakingKeeper.Delegate(ctx, addrs[0], sdk.TokensFromConsensusPower(powers), stakingtypes.Unbonded, val1, true)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return addrs[0], valAddrs[0]
}
