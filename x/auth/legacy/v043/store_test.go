package v043_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateVestingAccounts(t *testing.T) {
	testCases := []struct {
		name        string
		prepareFunc func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress)
		garbageFunc func(ctx sdk.Context, vesting exported.VestingAccount, app *simapp.SimApp) error
		tokenAmount int64
		expVested   int64
		expFree     int64
		blockTime   int64
	}{
		{
			"delayed vesting has vested, multiple delegations less than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			300,
			0,
			300,
			0,
		},
		{
			"delayed vesting has vested, single delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
			300,
			0,
			300,
			0,
		},
		{
			"delayed vesting has vested, multiple delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			300,
			0,
			300,
			0,
		},
		{
			"delayed vesting has not vested, single delegations  which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(200)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
			300,
			200,
			100,
			0,
		},
		{
			"delayed vesting has not vested, multiple delegations which exceed the vested amount",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			300,
			200,
			100,
			0,
		},
		{
			"not end time",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			300,
			300,
			0,
			0,
		},
		{
			"delayed vesting has not vested, single delegation greater than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(1, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
			300,
			300,
			0,
			0,
		},
		{
			"delayed vesting has vested, single delegation greater than the total account balance",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().Unix())

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
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
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
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
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
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
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))
				delayedAccount := types.NewContinuousVestingAccount(baseAccount, vestedCoins, startTime, endTime)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)
			},
			cleartTrackingFields,
			300,
			0,
			300,
			0,
		},
		{
			"periodic vesting account, yet to be vested, some rewards delegated",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			300,
			100,
			50,
			0,
		},
		{
			"periodic vesting account, nothing has vested yet",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				/*
					Test case:
					 - periodic vesting account starts at time 1601042400
					 - account balance and original vesting: 3666666670000
					 - nothing has vested, we put the block time slightly after start time
					 - expected vested: original vesting amount
					 - expected free: zero
					 - we're delegating the full original vesting
				*/
				startTime := int64(1601042400)
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			3666666670000,
			3666666670000,
			0,
			1601042400 + 1,
		},
		{
			"periodic vesting account, all has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				/*
					Test case:
					 - periodic vesting account starts at time 1601042400
					 - account balance and original vesting: 3666666670000
					 - all has vested, so we set the block time at initial time + sum of all periods times + 1 => 1601042400 + 31536000 + 15897600 + 15897600 + 1
					 - expected vested: zero
					 - expected free: original vesting amount
					 - we're delegating the full original vesting
				*/
				startTime := int64(1601042400)
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			3666666670000,
			0,
			3666666670000,
			1601042400 + 31536000 + 15897600 + 15897600 + 1,
		},
		{
			"periodic vesting account, first period has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				/*
					Test case:
					 - periodic vesting account starts at time 1601042400
					 - account balance and original vesting: 3666666670000
					 - first period have vested, so we set the block time at initial time + time of the first periods + 1 => 1601042400 + 31536000 + 1
					 - expected vested: original vesting - first period amount
					 - expected free: first period amount
					 - we're delegating the full original vesting
				*/
				startTime := int64(1601042400)
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			3666666670000,
			3666666670000 - 1833333335000,
			1833333335000,
			1601042400 + 31536000 + 1,
		},
		{
			"periodic vesting account, first 2 period has vested",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				/*
					Test case:
					 - periodic vesting account starts at time 1601042400
					 - account balance and original vesting: 3666666670000
					 - first 2 periods have vested, so we set the block time at initial time + time of the two periods + 1 => 1601042400 + 31536000 + 15638400 + 1
					 - expected vested: original vesting - (sum of the first two periods amounts)
					 - expected free: sum of the first two periods
					 - we're delegating the full original vesting
				*/
				startTime := int64(1601042400)
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
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
			cleartTrackingFields,
			3666666670000,
			3666666670000 - 1833333335000 - 916666667500,
			1833333335000 + 916666667500,
			1601042400 + 31536000 + 15638400 + 1,
		},
		{
			"vesting account has unbonding delegations in place",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))

				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(10, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)

				// delegation of the original vesting
				_, err := app.StakingKeeper.Delegate(ctx, delegatorAddr, sdk.NewInt(300), stakingtypes.Unbonded, validator, true)
				require.NoError(t, err)

				ctx = ctx.WithBlockTime(ctx.BlockTime().AddDate(1, 0, 0))

				valAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
				require.NoError(t, err)

				// un-delegation of the original vesting
				_, err = app.StakingKeeper.Undelegate(ctx, delegatorAddr, valAddr, sdk.NewDecFromInt(sdk.NewInt(300)))
				require.NoError(t, err)
			},
			cleartTrackingFields,
			450,
			300,
			0,
			0,
		},
		{
			"vesting account has never delegated anything",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))

				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(10, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)
			},
			cleartTrackingFields,
			450,
			0,
			0,
			0,
		},
		{
			"vesting account has no delegation but dirty DelegatedFree and DelegatedVesting fields",
			func(app *simapp.SimApp, ctx sdk.Context, validator stakingtypes.Validator, delegatorAddr sdk.AccAddress) {
				baseAccount := authtypes.NewBaseAccountWithAddress(delegatorAddr)
				vestedCoins := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(300)))

				delayedAccount := types.NewDelayedVestingAccount(baseAccount, vestedCoins, ctx.BlockTime().AddDate(10, 0, 0).Unix())

				app.AccountKeeper.SetAccount(ctx, delayedAccount)
			},
			dirtyTrackingFields,
			450,
			0,
			0,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			app := simapp.Setup(false)
			ctx := app.BaseApp.NewContext(false, tmproto.Header{
				Time: time.Now(),
			})

			addrs := simapp.AddTestAddrs(app, ctx, 1, sdk.NewInt(tc.tokenAmount))
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
			require.NoError(t, tc.garbageFunc(ctx, vestingAccount, app))

			m := authkeeper.NewMigrator(app.AccountKeeper, app.GRPCQueryRouter())
			require.NoError(t, m.Migrate1to2(ctx))

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

func trackingCorrected(ctx sdk.Context, t *testing.T, ak authkeeper.AccountKeeper, addr sdk.AccAddress, expDelVesting, expDelFree sdk.Coins) {
	t.Helper()
	baseAccount := ak.GetAccount(ctx, addr)
	vDA, ok := baseAccount.(exported.VestingAccount)
	require.True(t, ok)

	vestedOk := expDelVesting.IsEqual(vDA.GetDelegatedVesting())
	freeOk := expDelFree.IsEqual(vDA.GetDelegatedFree())
	require.True(t, vestedOk, vDA.GetDelegatedVesting().String())
	require.True(t, freeOk, vDA.GetDelegatedFree().String())
}

func cleartTrackingFields(ctx sdk.Context, vesting exported.VestingAccount, app *simapp.SimApp) error {
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

func dirtyTrackingFields(ctx sdk.Context, vesting exported.VestingAccount, app *simapp.SimApp) error {
	dirt := sdk.NewCoins(sdk.NewInt64Coin("stake", 42))

	switch t := vesting.(type) {
	case *types.DelayedVestingAccount:
		t.DelegatedFree = dirt
		t.DelegatedVesting = dirt
		app.AccountKeeper.SetAccount(ctx, t)
	case *types.ContinuousVestingAccount:
		t.DelegatedFree = dirt
		t.DelegatedVesting = dirt
		app.AccountKeeper.SetAccount(ctx, t)
	case *types.PeriodicVestingAccount:
		t.DelegatedFree = dirt
		t.DelegatedVesting = dirt
		app.AccountKeeper.SetAccount(ctx, t)
	default:
		return fmt.Errorf("expected vesting account, found %t", t)
	}

	return nil
}

func createValidator(t *testing.T, ctx sdk.Context, app *simapp.SimApp, powers int64) (sdk.AccAddress, sdk.ValAddress) {
	t.Helper()
	valTokens := sdk.TokensFromConsensusPower(powers, sdk.DefaultPowerReduction)
	addrs := simapp.AddTestAddrsIncremental(app, ctx, 1, valTokens)
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

	_, err = app.StakingKeeper.Delegate(ctx, addrs[0], valTokens, stakingtypes.Unbonded, val1, true)
	require.NoError(t, err)

	_ = staking.EndBlocker(ctx, app.StakingKeeper)

	return addrs[0], valAddrs[0]
}
