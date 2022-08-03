package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

func TestCancelUnbondingDelegation(t *testing.T) {
	// setup the app
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)
	bondDenom := app.StakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 5)

	require.NoError(t, testutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := app.BankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), app.StakingKeeper.BondDenom(ctx))
	require.Equal(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	delAddrs := simapp.AddTestAddrsIncremental(app, ctx, 2, sdk.NewInt(10000))
	validators := app.StakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, len(validators), 1)

	validatorAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	require.NoError(t, err)
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(app.StakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
	)

	// set and retrieve a record
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid coin",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin("dump_coin", sdk.NewInt(4)),
				CreationHeight:   0,
			},
		},
		{
			Name:      "validator not exists",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid delegator address",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
		},
		{
			Name:      "invalid amount",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			Name:      "success",
			ExceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, &testCase.req)
			if testCase.ExceptErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				balanceForNotBondedPool := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				require.Equal(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	msgServer := keeper.NewMsgServerImpl(app.StakingKeeper)

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "valid params",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params:    types.DefaultParams(),
			},
			expErr: false,
		},
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    types.DefaultParams(),
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "negative commission rate",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: math.LegacyNewDec(-10),
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be negative",
		},
		{
			name: "commission rate cannot be bigger than 100",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: math.LegacyNewDec(2),
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "minimum commission rate cannot be greater than 100%",
		},
		{
			name: "invalid bond denom",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         "",
				},
			},
			expErr:    true,
			expErrMsg: "bond denom cannot be blank",
		},
		{
			name: "max validators most be positive",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     0,
					MaxEntries:        types.DefaultMaxEntries,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max validators must be positive",
		},
		{
			name: "max entries most be positive",
			input: &types.MsgUpdateParams{
				Authority: app.StakingKeeper.GetAuthority(),
				Params: types.Params{
					MinCommissionRate: types.DefaultMinCommissionRate,
					UnbondingTime:     types.DefaultUnbondingTime,
					MaxValidators:     types.DefaultMaxValidators,
					MaxEntries:        0,
					HistoricalEntries: types.DefaultHistoricalEntries,
					BondDenom:         types.BondStatusBonded,
				},
			},
			expErr:    true,
			expErrMsg: "max entries must be positive",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.UpdateParams(ctx, tc.input)
			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
