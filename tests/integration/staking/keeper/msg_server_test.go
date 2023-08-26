package keeper_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCancelUnbondingDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	bondDenom, err := f.stakingKeeper.BondDenom(ctx)
	assert.NilError(t, err)

	// set the not bonded pool module account
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)
	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 5)

	assert.NilError(t, testutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := f.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom)
	assert.DeepEqual(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 2, math.NewInt(10000))
	valAddr := sdk.ValAddress(addrs[0])
	delegatorAddr := addrs[1]

	// setup a new validator with bonded status
	validator, err := types.NewValidator(valAddr.String(), PKs[0], types.NewDescription("Validator", "", "", "", ""))
	validator.Status = types.Bonded
	assert.NilError(t, err)
	assert.NilError(t, f.stakingKeeper.SetValidator(ctx, validator))

	validatorAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	assert.NilError(t, err)

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(bondDenom, 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
		0,
		address.NewBech32Codec("cosmosvaloper"), address.NewBech32Codec("cosmos"),
	)

	// set and retrieve a record
	assert.NilError(t, f.stakingKeeper.SetUnbondingDelegation(ctx, ubd))
	resUnbond, found := f.stakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	assert.Assert(t, found)
	assert.DeepEqual(t, ubd, resUnbond)

	testCases := []struct {
		name      string
		exceptErr bool
		req       types.MsgCancelUnbondingDelegation
		expErrMsg string
	}{
		{
			name:      "entry not found at height",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(bondDenom, math.NewInt(4)),
				CreationHeight:   11,
			},
			expErrMsg: "unbonding delegation entry is not found at block height",
		},
		{
			name:      "invalid height",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(bondDenom, math.NewInt(4)),
				CreationHeight:   0,
			},
			expErrMsg: "invalid height",
		},
		{
			name:      "invalid coin",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin("dump_coin", math.NewInt(4)),
				CreationHeight:   10,
			},
			expErrMsg: "invalid coin denomination",
		},
		{
			name:      "validator not exists",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: sdk.ValAddress(sdk.AccAddress("asdsad")).String(),
				Amount:           unbondingAmount,
				CreationHeight:   10,
			},
			expErrMsg: "validator does not exist",
		},
		{
			name:      "invalid delegator address",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: "invalid_delegator_addrtess",
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount,
				CreationHeight:   0,
			},
			expErrMsg: "decoding bech32 failed",
		},
		{
			name:      "invalid amount",
			exceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Add(sdk.NewInt64Coin(bondDenom, 10)),
				CreationHeight:   10,
			},
			expErrMsg: "amount is greater than the unbonding delegation entry balance",
		},
		{
			name:      "success",
			exceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1)),
				CreationHeight:   10,
			},
		},
		{
			name:      "success",
			exceptErr: false,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           unbondingAmount.Sub(unbondingAmount.Sub(sdk.NewInt64Coin(bondDenom, 1))),
				CreationHeight:   10,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, &tc.req)
			if tc.exceptErr {
				assert.ErrorContains(t, err, tc.expErrMsg)
			} else {
				assert.NilError(t, err)
				balanceForNotBondedPool := f.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom)
				assert.DeepEqual(t, balanceForNotBondedPool, moduleBalance.Sub(tc.req.Amount))
				moduleBalance = moduleBalance.Sub(tc.req.Amount)
			}
		})
	}
}
