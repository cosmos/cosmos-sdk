package keeper_test

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkaddress "github.com/cosmos/cosmos-sdk/types/address"
	accountkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
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

	assert.NilError(t, banktestutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
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

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := msgServer.CancelUnbondingDelegation(ctx, &testCase.req)
			if testCase.exceptErr {
				assert.ErrorContains(t, err, testCase.expErrMsg)
			} else {
				assert.NilError(t, err)
				balanceForNotBondedPool := f.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), bondDenom)
				assert.DeepEqual(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}

func TestTokenizeSharesAndRedeemTokens(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx

	stakingKeeper := f.stakingKeeper

	liquidStakingCapStrict := math.LegacyZeroDec()
	liquidStakingCapConservative := math.LegacyMustNewDecFromStr("0.8")
	liquidStakingCapDisabled := math.LegacyOneDec()

	validatorBondStrict := math.LegacyOneDec()
	validatorBondConservative := math.LegacyNewDec(10)
	validatorBondDisabled := math.LegacyNewDec(-1)

	testCases := []struct {
		name                          string
		vestingAmount                 math.Int
		delegationAmount              math.Int
		tokenizeShareAmount           math.Int
		redeemAmount                  math.Int
		targetVestingDelAfterShare    math.Int
		targetVestingDelAfterRedeem   math.Int
		globalLiquidStakingCap        math.LegacyDec
		slashFactor                   math.LegacyDec
		validatorLiquidStakingCap     math.LegacyDec
		validatorBondFactor           math.LegacyDec
		validatorBondDelegation       bool
		validatorBondDelegatorIndex   int
		delegatorIsLSTP               bool
		expTokenizeErr                bool
		expRedeemErr                  bool
		prevAccountDelegationExists   bool
		recordAccountDelegationExists bool
	}{
		{
			name:                          "full amount tokenize and redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "full amount tokenize and partial redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                          "partial amount tokenize and full redeem",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyZeroDec(),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   true,
			recordAccountDelegationExists: false,
		},
		{
			name:                          "tokenize and redeem with slash",
			vestingAmount:                 math.NewInt(0),
			delegationAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:           stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                   math.LegacyMustNewDecFromStr("0.1"),
			globalLiquidStakingCap:        liquidStakingCapDisabled,
			validatorLiquidStakingCap:     liquidStakingCapDisabled,
			validatorBondFactor:           validatorBondDisabled,
			validatorBondDelegation:       false,
			expTokenizeErr:                false,
			expRedeemErr:                  false,
			prevAccountDelegationExists:   false,
			recordAccountDelegationExists: true,
		},
		{
			name:                      "over tokenize",
			vestingAmount:             math.NewInt(0),
			delegationAmount:          stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       stakingKeeper.TokensFromConsensusPower(ctx, 30),
			redeemAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:               math.LegacyZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			validatorBondFactor:       validatorBondDisabled,
			validatorBondDelegation:   false,
			expTokenizeErr:            true,
			expRedeemErr:              false,
		},
		{
			name:                      "over redeem",
			vestingAmount:             math.NewInt(0),
			delegationAmount:          stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:       stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:              stakingKeeper.TokensFromConsensusPower(ctx, 40),
			slashFactor:               math.LegacyZeroDec(),
			globalLiquidStakingCap:    liquidStakingCapDisabled,
			validatorLiquidStakingCap: liquidStakingCapDisabled,
			validatorBondFactor:       validatorBondDisabled,
			validatorBondDelegation:   false,
			expTokenizeErr:            false,
			expRedeemErr:              true,
		},
		{
			name:                        "vesting account tokenize share failure",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 20),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 20),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     false,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "vesting account tokenize share success",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     false,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "try tokenize share for a validator-bond delegation",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 1,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict validator-bond - tokenization fails",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondStrict,
			validatorBondDelegation:     false,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative validator-bond - successful tokenization",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict global liquid staking cap - tokenization fails",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapStrict,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative global liquid staking cap - successful tokenization",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapDisabled,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "strict validator liquid staking cap - tokenization fails",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapStrict,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              true,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "conservative validator liquid staking cap - successful tokenization",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapDisabled,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondDisabled,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "all caps set conservatively - successful tokenize share",
			vestingAmount:               stakingKeeper.TokensFromConsensusPower(ctx, 10),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondConservative,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
		{
			name:                        "delegator is a liquid staking provider - accounting should not update",
			vestingAmount:               math.ZeroInt(),
			delegationAmount:            stakingKeeper.TokensFromConsensusPower(ctx, 20),
			tokenizeShareAmount:         stakingKeeper.TokensFromConsensusPower(ctx, 10),
			redeemAmount:                stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterShare:  stakingKeeper.TokensFromConsensusPower(ctx, 10),
			targetVestingDelAfterRedeem: stakingKeeper.TokensFromConsensusPower(ctx, 10),
			slashFactor:                 math.LegacyZeroDec(),
			globalLiquidStakingCap:      liquidStakingCapConservative,
			validatorLiquidStakingCap:   liquidStakingCapConservative,
			validatorBondFactor:         validatorBondConservative,
			delegatorIsLSTP:             true,
			validatorBondDelegation:     true,
			validatorBondDelegatorIndex: 0,
			expTokenizeErr:              false,
			expRedeemErr:                false,
			prevAccountDelegationExists: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			f := initFixture(t)

			ctx := f.sdkCtx
			var (
				bankKeeper    = f.bankKeeper
				accountKeeper = f.accountKeeper
				stakingKeeper = f.stakingKeeper
			)
			addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
			addrAcc1, addrAcc2 := addrs[0], addrs[1]
			addrVal1, addrVal2 := sdk.ValAddress(addrAcc1), sdk.ValAddress(addrAcc2)

			// Create ICA module account
			icaAccountAddress := createICAAccount(ctx, accountKeeper)

			// Fund module account
			bondDenom, err := stakingKeeper.BondDenom(ctx)
			require.NoError(t, err)
			delegationCoin := sdk.NewCoin(bondDenom, tc.delegationAmount)
			err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegationCoin))
			require.NoError(t, err)
			err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, icaAccountAddress, sdk.NewCoins(delegationCoin))
			require.NoError(t, err)

			// set the delegator address depending on whether the delegator should be a liquid staking provider
			delegatorAccount := addrAcc2
			if tc.delegatorIsLSTP {
				delegatorAccount = icaAccountAddress
			}

			// set validator bond factor and global liquid staking cap
			params, err := stakingKeeper.GetParams(ctx)
			require.NoError(t, err)
			params.ValidatorBondFactor = tc.validatorBondFactor
			params.GlobalLiquidStakingCap = tc.globalLiquidStakingCap
			params.ValidatorLiquidStakingCap = tc.validatorLiquidStakingCap
			stakingKeeper.SetParams(ctx, params)

			// set the total liquid staked tokens
			stakingKeeper.SetTotalLiquidStakedTokens(ctx, math.ZeroInt())

			if !tc.vestingAmount.IsZero() {
				// create vesting account
				acc2 := accountKeeper.GetAccount(ctx, addrAcc2).(*authtypes.BaseAccount)
				initialVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, tc.vestingAmount))
				baseVestingWithCoins, err := vestingtypes.NewBaseVestingAccount(acc2, initialVesting, time.Now().Unix()+86400*365)
				require.NoError(t, err)
				delayedVestingAccount := vestingtypes.NewDelayedVestingAccountRaw(baseVestingWithCoins)
				accountKeeper.SetAccount(ctx, delayedVestingAccount)
			}

			pubKeys := simtestutil.CreateTestPubKeys(2)
			pk1, pk2 := pubKeys[0], pubKeys[1]

			// Create Validators and Delegation
			val1 := testutil.NewValidator(t, addrVal1, pk1)
			val1.Status = types.Bonded
			stakingKeeper.SetValidator(ctx, val1)
			stakingKeeper.SetValidatorByPowerIndex(ctx, val1)
			err = stakingKeeper.SetValidatorByConsAddr(ctx, val1)
			require.NoError(t, err)

			val2 := testutil.NewValidator(t, addrVal2, pk2)
			val2.Status = types.Bonded
			stakingKeeper.SetValidator(ctx, val2)
			stakingKeeper.SetValidatorByPowerIndex(ctx, val2)
			err = stakingKeeper.SetValidatorByConsAddr(ctx, val2)
			require.NoError(t, err)

			// Delegate from both the main delegator as well as a random account so there is a
			// non-zero delegation after redemption
			err = delegateCoinsFromAccount(ctx, *stakingKeeper, delegatorAccount, tc.delegationAmount, val1)
			require.NoError(t, err)

			// apply TM updates
			applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

			_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.NoError(t, err, "delegation not found after delegate")

			lastRecordID := stakingKeeper.GetLastTokenizeShareRecordID(ctx)
			oldValidator, err := stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)

			msgServer := keeper.NewMsgServerImpl(stakingKeeper)
			if tc.validatorBondDelegation {
				err := delegateCoinsFromAccount(ctx, *stakingKeeper, addrs[tc.validatorBondDelegatorIndex], tc.delegationAmount, val1)
				require.NoError(t, err)
				_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
					DelegatorAddress: addrs[tc.validatorBondDelegatorIndex].String(),
					ValidatorAddress: addrVal1.String(),
				})
				require.NoError(t, err)
			}

			resp, err := msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
				DelegatorAddress:    delegatorAccount.String(),
				ValidatorAddress:    addrVal1.String(),
				Amount:              sdk.NewCoin(bondDenom, tc.tokenizeShareAmount),
				TokenizedShareOwner: delegatorAccount.String(),
			})
			if tc.expTokenizeErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// check last record id increase
			require.Equal(t, lastRecordID+1, stakingKeeper.GetLastTokenizeShareRecordID(ctx))

			// ensure validator's total tokens is consistent
			newValidator, err := stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			// if the delegator was not a provider, check that the total liquid staked and validator liquid shares increased
			totalLiquidTokensAfterTokenization := stakingKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterTokenization := newValidator.LiquidShares
			if !tc.delegatorIsLSTP {
				require.Equal(t, tc.tokenizeShareAmount.String(), totalLiquidTokensAfterTokenization.String(), "total liquid tokens after tokenization")
				require.Equal(t, tc.tokenizeShareAmount.String(), validatorLiquidSharesAfterTokenization.TruncateInt().String(), "validator liquid shares after tokenization")
			} else {
				require.True(t, totalLiquidTokensAfterTokenization.IsZero(), "zero liquid tokens after tokenization")
				require.True(t, validatorLiquidSharesAfterTokenization.IsZero(), "zero liquid validator shares after tokenization")
			}

			if tc.vestingAmount.IsPositive() {
				acc := accountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(bondDenom).String(), tc.targetVestingDelAfterShare.String())
			}

			if tc.prevAccountDelegationExists {
				_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.NoError(t, err, "delegation not found after partial tokenize share")
			} else {
				_, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
				require.ErrorIs(t, err, types.ErrNoDelegation, "delegation found after full tokenize share")
			}

			shareToken := bankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, resp.Amount, shareToken)
			_, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err, "validator not found")

			records := stakingKeeper.GetAllTokenizeShareRecords(ctx)
			require.Len(t, records, 1)
			delegation, err := stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
			require.NoError(t, err, "delegation not found from tokenize share module account after tokenize share")

			// slash before redeem
			slashedTokens := math.ZeroInt()
			redeemedShares := tc.redeemAmount
			redeemedTokens := tc.redeemAmount
			if tc.slashFactor.IsPositive() {
				consAddr, err := val1.GetConsAddr()
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(100)
				val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
				require.NoError(t, err)
				power := stakingKeeper.TokensToConsensusPower(ctx, val1.Tokens)
				stakingKeeper.Slash(ctx, consAddr, 10, power, tc.slashFactor)
				slashedTokens = math.LegacyNewDecFromInt(val1.Tokens).Mul(tc.slashFactor).TruncateInt()

				val1, _ := stakingKeeper.GetValidator(ctx, addrVal1)
				redeemedTokens = val1.TokensFromShares(math.LegacyNewDecFromInt(redeemedShares)).TruncateInt()
			}

			// get delegator balance and delegation
			bondDenomAmountBefore := bankKeeper.GetBalance(ctx, delegatorAccount, bondDenom)
			val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if errors.Is(err, types.ErrNoDelegation) {
				delegation = types.Delegation{Shares: math.LegacyZeroDec()}
			}
			delAmountBefore := val1.TokensFromShares(delegation.Shares)
			oldValidator, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)

			_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
				DelegatorAddress: delegatorAccount.String(),
				Amount:           sdk.NewCoin(resp.Amount.Denom, tc.redeemAmount),
			})
			if tc.expRedeemErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// ensure validator's total tokens is consistent
			newValidator, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			require.Equal(t, oldValidator.Tokens, newValidator.Tokens)

			// if the delegator was not a liquid staking provider, check that the total liquid staked
			// and liquid shares decreased
			totalLiquidTokensAfterRedemption := stakingKeeper.GetTotalLiquidStakedTokens(ctx)
			validatorLiquidSharesAfterRedemption := newValidator.LiquidShares
			expectedLiquidTokens := totalLiquidTokensAfterTokenization.Sub(redeemedTokens).Sub(slashedTokens)
			expectedLiquidShares := validatorLiquidSharesAfterTokenization.Sub(math.LegacyNewDecFromInt(redeemedShares))
			if !tc.delegatorIsLSTP {
				require.Equal(t, expectedLiquidTokens.String(), totalLiquidTokensAfterRedemption.String(), "total liquid tokens after redemption")
				require.Equal(t, expectedLiquidShares.String(), validatorLiquidSharesAfterRedemption.String(), "validator liquid shares after tokenization")
			} else {
				require.True(t, totalLiquidTokensAfterRedemption.IsZero(), "zero liquid tokens after redemption")
				require.True(t, validatorLiquidSharesAfterRedemption.IsZero(), "zero liquid validator shares after redemption")
			}

			if tc.vestingAmount.IsPositive() {
				acc := accountKeeper.GetAccount(ctx, addrAcc2)
				vestingAcc := acc.(vesting.VestingAccount)
				require.Equal(t, vestingAcc.GetDelegatedVesting().AmountOf(bondDenom).String(), tc.targetVestingDelAfterRedeem.String())
			}

			expectedDelegatedShares := math.LegacyNewDecFromInt(tc.delegationAmount.Sub(tc.tokenizeShareAmount).Add(tc.redeemAmount))
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			require.NoError(t, err, "delegation not found after redeem tokens")
			require.Equal(t, delegatorAccount.String(), delegation.DelegatorAddress)
			require.Equal(t, addrVal1.String(), delegation.ValidatorAddress)
			require.Equal(t, expectedDelegatedShares, delegation.Shares, "delegation shares after redeem")

			// check delegator balance is not changed
			bondDenomAmountAfter := bankKeeper.GetBalance(ctx, delegatorAccount, bondDenom)
			require.Equal(t, bondDenomAmountAfter.Amount.String(), bondDenomAmountBefore.Amount.String())

			// get delegation amount is changed correctly
			val1, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err)
			delegation, err = stakingKeeper.GetDelegation(ctx, delegatorAccount, addrVal1)
			if errors.Is(err, types.ErrNoDelegation) {
				delegation = types.Delegation{Shares: math.LegacyZeroDec()}
			}
			delAmountAfter := val1.TokensFromShares(delegation.Shares)
			require.Equal(t, delAmountAfter.String(), delAmountBefore.Add(math.LegacyNewDecFromInt(tc.redeemAmount).Mul(math.LegacyOneDec().Sub(tc.slashFactor))).String())

			shareToken = bankKeeper.GetBalance(ctx, delegatorAccount, resp.Amount.Denom)
			require.Equal(t, shareToken.Amount.String(), tc.tokenizeShareAmount.Sub(tc.redeemAmount).String())
			_, err = stakingKeeper.GetValidator(ctx, addrVal1)
			require.NoError(t, err, "validator not found")

			if tc.recordAccountDelegationExists {
				_, err = stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.NoError(t, err, "delegation not found from tokenize share module account after redeem partial amount")

				records = stakingKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 1)
			} else {
				_, err = stakingKeeper.GetDelegation(ctx, records[0].GetModuleAddress(), addrVal1)
				require.True(t, errors.Is(err, types.ErrNoDelegation), "delegation found from tokenize share module account after redeem full amount")

				records = stakingKeeper.GetAllTokenizeShareRecords(ctx)
				require.Len(t, records, 0)
			}
		})
	}
}

func TestRedelegationTokenization(t *testing.T) {
	// Test that a delegator with ongoing redelegation cannot
	// tokenize any shares until the redelegation is complete.
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	alice := addrs[0]

	validatorAAddress := sdk.ValAddress(addrs[1])
	val1 := testutil.NewValidator(t, validatorAAddress, pk1)
	val1.Status = types.Bonded
	stakingKeeper.SetValidator(ctx, val1)
	stakingKeeper.SetValidatorByPowerIndex(ctx, val1)
	err := stakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	_, validatorBAddress := setupTestTokenizeAndRedeemConversion(t, *stakingKeeper, bankKeeper, ctx)

	delegateAmount := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	bondedDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondedDenom, delegateAmount)

	// Alice delegates to validatorA
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: alice.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err)

	// Alice redelegates to validatorB
	redelegateAmount := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	redelegateCoin := sdk.NewCoin(bondedDenom, redelegateAmount)
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    alice.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorBAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err)

	redelegation, err := stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 1, "expect one redelegation")
	require.Len(t, redelegation[0].Entries, 1, "expect one redelegation entry")

	// Alice attempts to tokenize the redelegation, but this fails because the redelegation is ongoing
	tokenizedAmount := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	tokenizedCoin := sdk.NewCoin(bondedDenom, tokenizedAmount)
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    alice.String(),
		ValidatorAddress:    validatorBAddress.String(),
		Amount:              tokenizedCoin,
		TokenizedShareOwner: alice.String(),
	})
	require.Error(t, err)
	require.Equal(t, types.ErrRedelegationInProgress, err)

	// Check that the redelegation is still present
	redelegation, err = stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 1, "expect one redelegation")
	require.Len(t, redelegation[0].Entries, 1, "expect one redelegation entry")

	// advance time until the redelegations should mature
	// end block
	f.stakingKeeper.EndBlocker(ctx)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// advance by 22 days
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(22 * 24 * time.Hour))
	headerInfo := ctx.HeaderInfo()
	headerInfo.Time = ctx.BlockHeader().Time
	headerInfo.Height = ctx.BlockHeader().Height
	ctx = ctx.WithHeaderInfo(headerInfo)
	// begin block
	f.stakingKeeper.BeginBlocker(ctx)
	// end block
	f.stakingKeeper.EndBlocker(ctx)

	// check that the redelegation is removed
	redelegation, err = stakingKeeper.GetRedelegations(ctx, alice, uint16(10))
	require.NoError(t, err)
	require.Len(t, redelegation, 0, "expect no redelegations")

	// Alice attempts to tokenize the redelegation again, and this time it should succeed
	// because there is no ongoing redelegation
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    alice.String(),
		ValidatorAddress:    validatorBAddress.String(),
		Amount:              tokenizedCoin,
		TokenizedShareOwner: alice.String(),
	})
	require.NoError(t, err)

	// Check that the tokenization was successful
	shareRecord, err := stakingKeeper.GetTokenizeShareRecord(ctx, stakingKeeper.GetLastTokenizeShareRecordID(ctx))
	require.NoError(t, err, "expect to find token share record")
	require.Equal(t, alice.String(), shareRecord.Owner)
	require.Equal(t, validatorBAddress.String(), shareRecord.Validator)
}

// Helper function to setup a delegator and validator for the Tokenize/Redeem conversion tests
func setupTestTokenizeAndRedeemConversion(
	t *testing.T,
	sk keeper.Keeper,
	bk bankkeeper.Keeper,
	ctx sdk.Context,
) (delAddress sdk.AccAddress, valAddress sdk.ValAddress) {
	addresses := simtestutil.AddTestAddrs(bk, sk, ctx, 2, math.NewInt(1_000_000))

	pubKeys := simtestutil.CreateTestPubKeys(1)

	delegatorAddress := addresses[0]
	validatorAddress := sdk.ValAddress(addresses[1])

	validator, err := types.NewValidator(validatorAddress.String(), pubKeys[0], types.Description{})
	require.NoError(t, err)
	validator.DelegatorShares = math.LegacyNewDec(1_000_000)
	validator.Tokens = math.NewInt(1_000_000)
	validator.LiquidShares = math.LegacyNewDec(0)
	validator.Status = types.Bonded

	sk.SetValidator(ctx, validator)
	sk.SetValidatorByConsAddr(ctx, validator)

	return delegatorAddress, validatorAddress
}

// Simulate a slash by decrementing the validator's tokens
// We'll do this in a way such that the exchange rate is not an even integer
// and the shares associated with a delegation will have a long decimal
func simulateSlashWithImprecision(t *testing.T, sk keeper.Keeper, ctx sdk.Context, valAddress sdk.ValAddress) {
	validator, err := sk.GetValidator(ctx, valAddress)
	require.NoError(t, err)

	slashMagnitude := math.LegacyMustNewDecFromStr("0.1111111111")
	slashTokens := math.LegacyNewDecFromInt(validator.Tokens).Mul(slashMagnitude).TruncateInt()
	validator.Tokens = validator.Tokens.Sub(slashTokens)

	sk.SetValidator(ctx, validator)
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Slash -> Delegate -> Tokenize -> Redeem
// Note, in this example, there 2 tokens are lost during the decimal to int conversion
// during the unbonding step within tokenization and redemption
func TestTokenizeAndRedeemConversion_SlashBeforeDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *stakingKeeper, bankKeeper, ctx)

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	delegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// Confirm the number of shareTokens equals the number of shares truncated
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegation.Shares.TruncateInt().Int64() - 1 // 1 token was lost during unbonding
	require.Equal(t, expectedShareTokens, shareToken.Amount.Int64(), "share token amount")

	// Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm (almost) the full delegation was recovered - minus the 2 tokens from the precision error
	// (1 occurs during tokenization, and 1 occurs during redemption)
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegateAmount.Int64() - 2
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Slash -> Tokenize -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within tokenization
func TestTokenizeAndRedeemConversion_SlashBeforeTokenization(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *stakingKeeper, bankKeeper, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, err = stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Tokenize the new amount after the slash
	delegationAmountAfterSlash := validator.TokensFromShares(math.LegacyNewDecFromInt(delegateAmount)).TruncateInt()
	tokenizationCoin := sdk.NewCoin(bondDenom, delegationAmountAfterSlash)

	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              tokenizationCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up with the **new** number of shares associated
	// with the original delegated amount
	// Note: 1 token is lost during unbonding due to rounding
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens, err := validator.SharesFromTokens(tokenizationCoin.Amount)
	require.NoError(t, err)
	require.Equal(t, expectedShareTokens.TruncateInt().Int64()-1, shareToken.Amount.Int64(), "share token amount")

	// // Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the full tokenization amount was recovered - minus the 1 token from the precision error
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	expectedDelegationTokens := delegationAmountAfterSlash.Int64() - 1
	require.Equal(t, expectedDelegationTokens, endDelegationTokens, "final delegation tokens")
}

// Tests the conversion from tokenization and redemption from the following scenario:
// Delegate -> Tokenize -> Slash -> Redeem
// Note, in this example, there 1 token lost during the decimal to int conversion
// during the unbonding step within redemption
func TestTokenizeAndRedeemConversion_SlashBeforeRedemption(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	delegatorAddress, validatorAddress := setupTestTokenizeAndRedeemConversion(t, *stakingKeeper, bankKeeper, ctx)

	// Delegate and confirm the delegation record was created
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	})
	require.NoError(t, err, "no error expected when delegating")

	_, err = stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// Tokenize the full delegation amount
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              delegateCoin,
		TokenizedShareOwner: delegatorAddress.String(),
	})
	require.NoError(t, err, "no error expected when tokenizing")

	// The number of share tokens should line up 1:1 with the number of issued shares
	// Since the validator has not been slashed, the shares also line up 1;1
	// with the original delegation amount
	shareDenom := validatorAddress.String() + "/1"
	shareToken := bankKeeper.GetBalance(ctx, delegatorAddress, shareDenom)
	expectedShareTokens := delegateAmount
	require.Equal(t, expectedShareTokens.Int64(), shareToken.Amount.Int64(), "share token amount")

	// slash the validator
	simulateSlashWithImprecision(t, *stakingKeeper, ctx, validatorAddress)
	validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err)

	// Redeem the share tokens
	_, err = msgServer.RedeemTokensForShares(ctx, &types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
		Amount:           shareToken,
	})
	require.NoError(t, err, "no error expected when redeeming")

	// Confirm the original delegation, minus the slash, was recovered
	// There's an additional 1 token lost from precision error during unbonding
	delegationAmountAfterSlash := validator.TokensFromShares(math.LegacyNewDecFromInt(delegateAmount)).TruncateInt().Int64()
	newDelegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
	require.NoError(t, err)

	endDelegationTokens := validator.TokensFromShares(newDelegation.Shares).TruncateInt().Int64()
	require.Equal(t, delegationAmountAfterSlash-1, endDelegationTokens, "final delegation tokens")
}

func TestTransferTokenizeShareRecord(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)
	app, err := simtestutil.Setup(
		depinject.Configs(testutil.AppConfig, depinject.Supply(log.NewNopLogger())),
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false)
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 3, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1, addrAcc2, valAcc := addrs[0], addrs[1], addrs[2]
	addrVal := sdk.ValAddress(valAcc)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk := pubKeys[0]

	val, err := types.NewValidator(addrVal.String(), pk, types.Description{})
	require.NoError(t, err)

	stakingKeeper.SetValidator(ctx, val)
	stakingKeeper.SetValidatorByPowerIndex(ctx, val)

	// apply TM updates
	applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

	err = stakingKeeper.AddTokenizeShareRecord(ctx, types.TokenizeShareRecord{
		Id:            1,
		Owner:         addrAcc1.String(),
		ModuleAccount: "module_account",
		Validator:     val.String(),
	})
	require.NoError(t, err)

	_, err = msgServer.TransferTokenizeShareRecord(ctx, &types.MsgTransferTokenizeShareRecord{
		TokenizeShareRecordId: 1,
		Sender:                addrAcc1.String(),
		NewOwner:              addrAcc2.String(),
	})
	require.NoError(t, err)

	record, err := stakingKeeper.GetTokenizeShareRecord(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, record.Owner, addrAcc2.String())

	records := stakingKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc1)
	require.Len(t, records, 0)
	records = stakingKeeper.GetTokenizeShareRecordsByOwner(ctx, addrAcc2)
	require.Len(t, records, 1)
}

func TestValidatorBond(t *testing.T) {
	testCases := []struct {
		name                 string
		createValidator      bool
		createDelegation     bool
		alreadyValidatorBond bool
		delegatorIsLSTP      bool
		expectedErr          error
	}{
		{
			name:                 "successful validator bond",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
		},
		{
			name:                 "successful with existing validator bond",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: true,
			delegatorIsLSTP:      false,
		},
		{
			name:                 "validator does not not exist",
			createValidator:      false,
			createDelegation:     false,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
			expectedErr:          types.ErrNoValidatorFound,
		},
		{
			name:                 "delegation not exist case",
			createValidator:      true,
			createDelegation:     false,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      false,
			expectedErr:          types.ErrNoDelegation,
		},
		{
			name:                 "delegator is a liquid staking provider",
			createValidator:      true,
			createDelegation:     true,
			alreadyValidatorBond: false,
			delegatorIsLSTP:      true,
			expectedErr:          types.ErrValidatorBondNotAllowedFromModuleAccount,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc := tc
			t.Parallel()
			f := initFixture(t)

			ctx := f.sdkCtx
			var (
				accountKeeper = f.accountKeeper
				stakingKeeper = f.stakingKeeper
				bankKeeper    = f.bankKeeper
			)
			pubKeys := simtestutil.CreateTestPubKeys(2)

			validatorPubKey := pubKeys[0]
			delegatorPubKey := pubKeys[1]

			delegatorAddress := sdk.AccAddress(delegatorPubKey.Address())
			validatorAddress := sdk.ValAddress(validatorPubKey.Address())
			icaAccountAddress := createICAAccount(ctx, accountKeeper)

			// Set the delegator address to either be a user account or an ICA account depending on the test case
			if tc.delegatorIsLSTP {
				delegatorAddress = icaAccountAddress
			}

			// Fund the delegator
			delegationAmount := stakingKeeper.TokensFromConsensusPower(ctx, 20)
			bondDenom, err := stakingKeeper.BondDenom(ctx)
			require.NoError(t, err)
			coins := sdk.NewCoins(sdk.NewCoin(bondDenom, delegationAmount))

			err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, coins)
			require.NoError(t, err, "no error expected when minting")

			err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, delegatorAddress, coins)
			require.NoError(t, err, "no error expected when funding account")

			// Create Validator and delegation
			if tc.createValidator {
				validator, err := types.NewValidator(validatorAddress.String(), validatorPubKey, types.Description{})
				require.NoError(t, err)
				validator.Status = types.Bonded
				stakingKeeper.SetValidator(ctx, validator)
				stakingKeeper.SetValidatorByPowerIndex(ctx, validator)
				err = stakingKeeper.SetValidatorByConsAddr(ctx, validator)
				require.NoError(t, err)

				// Optionally create the delegation, depending on the test case
				if tc.createDelegation {
					_, err = stakingKeeper.Delegate(ctx, delegatorAddress, delegationAmount, types.Unbonded, validator, true)
					require.NoError(t, err, "no error expected when delegating")

					// Optionally, convert the delegation into a validator bond
					if tc.alreadyValidatorBond {
						delegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
						require.NoError(t, err, "delegation should have been found")

						delegation.ValidatorBond = true
						stakingKeeper.SetDelegation(ctx, delegation)
					}
				}
			}

			// Call ValidatorBond
			msgServer := keeper.NewMsgServerImpl(stakingKeeper)
			_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
				DelegatorAddress: delegatorAddress.String(),
				ValidatorAddress: validatorAddress.String(),
			})

			if tc.expectedErr != nil {
				require.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				require.NoError(t, err, "no error expected from validator bond transaction")

				// check validator bond true
				delegation, err := stakingKeeper.GetDelegation(ctx, delegatorAddress, validatorAddress)
				require.NoError(t, err, "delegation should have been found after validator bond")
				require.True(t, delegation.ValidatorBond, "delegation should be marked as a validator bond")

				// check validator bond shares
				validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
				require.NoError(t, err, "validator should have been found after validator bond")

				if tc.alreadyValidatorBond {
					require.True(t, validator.ValidatorBondShares.IsZero(), "validator bond shares should still be zero")
				} else {
					require.Equal(t, delegation.Shares.String(), validator.ValidatorBondShares.String(),
						"validator total shares should have increased")
				}
			}
		})
	}
}

func TestChangeValidatorBond(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)

	checkValidatorBondShares := func(validatorAddress sdk.ValAddress, expectedShares math.Int) {
		validator, err := stakingKeeper.GetValidator(ctx, validatorAddress)
		require.NoError(t, err, "validator should have been found")
		require.Equal(t, expectedShares.Int64(), validator.ValidatorBondShares.TruncateInt64(), "validator bond shares")
	}

	// Create a delegator and 3 validators
	addresses := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 4, math.NewInt(1_000_000))
	pubKeys := simtestutil.CreateTestPubKeys(4)

	validatorAPubKey := pubKeys[1]
	validatorBPubKey := pubKeys[2]
	validatorCPubKey := pubKeys[3]

	delegatorAddress := addresses[0]
	validatorAAddress := sdk.ValAddress(validatorAPubKey.Address())
	validatorBAddress := sdk.ValAddress(validatorBPubKey.Address())
	validatorCAddress := sdk.ValAddress(validatorCPubKey.Address())

	validatorA, err := types.NewValidator(validatorAAddress.String(), validatorAPubKey, types.Description{})
	require.NoError(t, err)
	validatorB, err := types.NewValidator(validatorBAddress.String(), validatorBPubKey, types.Description{})
	require.NoError(t, err)
	validatorC, err := types.NewValidator(validatorCAddress.String(), validatorCPubKey, types.Description{})
	require.NoError(t, err)

	validatorA.Tokens = math.NewInt(1_000_000)
	validatorB.Tokens = math.NewInt(1_000_000)
	validatorC.Tokens = math.NewInt(1_000_000)
	validatorA.DelegatorShares = math.LegacyNewDec(1_000_000)
	validatorB.DelegatorShares = math.LegacyNewDec(1_000_000)
	validatorC.DelegatorShares = math.LegacyNewDec(1_000_000)

	stakingKeeper.SetValidator(ctx, validatorA)
	stakingKeeper.SetValidator(ctx, validatorB)
	stakingKeeper.SetValidator(ctx, validatorC)

	// The test will go through Delegate/Redelegate/Undelegate messages with the following
	delegation1Amount := math.NewInt(1000)
	delegation2Amount := math.NewInt(1000)
	redelegateAmount := math.NewInt(500)
	undelegateAmount := math.NewInt(500)

	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegate1Coin := sdk.NewCoin(bondDenom, delegation1Amount)
	delegate2Coin := sdk.NewCoin(bondDenom, delegation2Amount)
	redelegateCoin := sdk.NewCoin(bondDenom, redelegateAmount)
	undelegateCoin := sdk.NewCoin(bondDenom, undelegateAmount)

	// Delegate to validator's A and C - validator bond shares should not change
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate1Coin,
	})
	require.NoError(t, err, "no error expected during first delegation")

	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorCAddress.String(),
		Amount:           delegate1Coin,
	})
	require.NoError(t, err, "no error expected during first delegation")

	checkValidatorBondShares(validatorAAddress, math.ZeroInt())
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, math.ZeroInt())

	// Flag the the delegations to validator A and C validator bond's
	// Their bond shares should increase
	_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
	})
	require.NoError(t, err, "no error expected during validator bond")

	_, err = msgServer.ValidatorBond(ctx, &types.MsgValidatorBond{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorCAddress.String(),
	})
	require.NoError(t, err, "no error expected during validator bond")

	checkValidatorBondShares(validatorAAddress, delegation1Amount)
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Delegate more to validator A - it should increase the validator bond shares
	_, err = msgServer.Delegate(ctx, &types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           delegate2Coin,
	})
	require.NoError(t, err, "no error expected during second delegation")

	checkValidatorBondShares(validatorAAddress, delegation1Amount.Add(delegation2Amount))
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Redelegate partially from A to B (where A is a validator bond and B is not)
	// It should remove the bond shares from A, and B's validator bond shares should not change
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorBAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	expectedBondSharesA := delegation1Amount.Add(delegation2Amount).Sub(redelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount)

	// Now redelegate from B to C (where B is not a validator bond, but C is)
	// Validator B's bond shares should remain at zero, but C's bond shares should increase
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorBAddress.String(),
		ValidatorDstAddress: validatorCAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, delegation1Amount.Add(redelegateAmount))

	// Redelegate partially from A to C (where C is a validator bond delegation)
	// It should remove the bond shares from A, and increase the bond shares on validator C
	_, err = msgServer.BeginRedelegate(ctx, &types.MsgBeginRedelegate{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorSrcAddress: validatorAAddress.String(),
		ValidatorDstAddress: validatorCAddress.String(),
		Amount:              redelegateCoin,
	})
	require.NoError(t, err, "no error expected during redelegation")

	expectedBondSharesA = expectedBondSharesA.Sub(redelegateAmount)
	expectedBondSharesC := delegation1Amount.Add(redelegateAmount).Add(redelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, expectedBondSharesC)

	// Undelegate from validator A - it should remove shares
	_, err = msgServer.Undelegate(ctx, &types.MsgUndelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAAddress.String(),
		Amount:           undelegateCoin,
	})
	require.NoError(t, err, "no error expected during undelegation")

	expectedBondSharesA = expectedBondSharesA.Sub(undelegateAmount)
	checkValidatorBondShares(validatorAAddress, expectedBondSharesA)
	checkValidatorBondShares(validatorBAddress, math.ZeroInt())
	checkValidatorBondShares(validatorCAddress, expectedBondSharesC)
}

func TestEnableDisableTokenizeShares(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)
	// Create a delegator and validator
	stakeAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	stakeToken := sdk.NewCoin(bondDenom, stakeAmount)

	addresses := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakeAmount)
	delegatorAddress := addresses[0]

	pubKeys := simtestutil.CreateTestPubKeys(1)
	validatorAddress := sdk.ValAddress(addresses[1])
	validator, err := types.NewValidator(validatorAddress.String(), pubKeys[0], types.Description{})
	require.NoError(t, err)

	validator.DelegatorShares = math.LegacyNewDec(1_000_000)
	validator.Tokens = math.NewInt(1_000_000)
	validator.Status = types.Bonded
	stakingKeeper.SetValidator(ctx, validator)

	// Fix block time and set unbonding period to 1 day
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	ctx = ctx.WithBlockTime(blockTime)

	unbondingPeriod := time.Hour * 24
	params, err := stakingKeeper.GetParams(ctx)
	require.NoError(t, err)
	params.UnbondingTime = unbondingPeriod
	stakingKeeper.SetParams(ctx, params)
	unlockTime := blockTime.Add(unbondingPeriod)

	// Build test messages (some of which will be reused)
	delegateMsg := types.MsgDelegate{
		DelegatorAddress: delegatorAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           stakeToken,
	}
	tokenizeMsg := types.MsgTokenizeShares{
		DelegatorAddress:    delegatorAddress.String(),
		ValidatorAddress:    validatorAddress.String(),
		Amount:              stakeToken,
		TokenizedShareOwner: delegatorAddress.String(),
	}
	redeemMsg := types.MsgRedeemTokensForShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	disableMsg := types.MsgDisableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}
	enableMsg := types.MsgEnableTokenizeShares{
		DelegatorAddress: delegatorAddress.String(),
	}

	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	// Delegate normally
	_, err = msgServer.Delegate(ctx, &delegateMsg)
	require.NoError(t, err, "no error expected when delegating")

	// Tokenize shares - it should succeed
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing shares for the first time")

	liquidToken := bankKeeper.GetBalance(ctx, delegatorAddress, validatorAddress.String()+"/1")
	require.Equal(t, stakeAmount.Int64(), liquidToken.Amount.Int64(), "user received token after tokenizing share")

	// Redeem to remove all tokenized shares
	redeemMsg.Amount = liquidToken
	_, err = msgServer.RedeemTokensForShares(ctx, &redeemMsg)
	require.NoError(t, err, "no error expected when redeeming")

	// Attempt to enable tokenizing shares when there is no lock in place, it should error
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesAlreadyEnabledForAccount)

	// Attempt to disable when no lock is in place, it should succeed
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.NoError(t, err, "no error expected when disabling tokenization")

	// Disabling again while the lock is already in place, should error
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesAlreadyDisabledForAccount)

	// Attempt to tokenize, it should fail since tokenization is disabled
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesDisabledForAccount)

	// Now enable tokenization
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization")

	// Attempt to tokenize again, it should still fail since the unbonding period has
	// not passed and the lock is still active
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.ErrorIs(t, err, types.ErrTokenizeSharesDisabledForAccount)
	require.ErrorContains(t, err, fmt.Sprintf("tokenization will be allowed at %s",
		blockTime.Add(unbondingPeriod)))

	// Confirm the unlock is queued
	authorizations := stakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Equal(t, []string{delegatorAddress.String()}, authorizations.Addresses,
		"pending tokenize share authorizations")

	// Disable tokenization again - it should remove the pending record from the queue
	_, err = msgServer.DisableTokenizeShares(ctx, &disableMsg)
	require.NoError(t, err, "no error expected when re-enabling tokenization")

	authorizations = stakingKeeper.GetPendingTokenizeShareAuthorizations(ctx, unlockTime)
	require.Empty(t, authorizations.Addresses, "there should be no pending authorizations in the queue")

	// Enable one more time
	_, err = msgServer.EnableTokenizeShares(ctx, &enableMsg)
	require.NoError(t, err, "no error expected when enabling tokenization again")

	// Increment the block time by the unbonding period and remove the expired locks
	ctx = ctx.WithBlockTime(unlockTime)
	stakingKeeper.RemoveExpiredTokenizeShareLocks(ctx, ctx.BlockTime())

	// Attempt to tokenize again, it should succeed this time since the lock has expired
	_, err = msgServer.TokenizeShares(ctx, &tokenizeMsg)
	require.NoError(t, err, "no error expected when tokenizing after lock has expired")
}

func TestUnbondValidator(t *testing.T) {
	var (
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)

	app, err := simtestutil.Setup(
		depinject.Configs(testutil.AppConfig, depinject.Supply(log.NewNopLogger())),
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)
	ctx := app.BaseApp.NewContext(false)

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 2, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1 := addrs[0]
	addrVal1 := sdk.ValAddress(addrAcc1)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	val1, err := types.NewValidator(addrVal1.String(), pk1, types.Description{})
	require.NoError(t, err)
	val1.Status = types.Bonded
	stakingKeeper.SetValidator(ctx, val1)
	stakingKeeper.SetValidatorByPowerIndex(ctx, val1)
	err = stakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	// try unbonding not available validator
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	_, err = msgServer.UnbondValidator(ctx, &types.MsgUnbondValidator{
		ValidatorAddress: addrs[1].String(),
	})
	require.Error(t, err)

	// unbond validator
	_, err = msgServer.UnbondValidator(ctx, &types.MsgUnbondValidator{
		ValidatorAddress: addrAcc1.String(),
	})
	require.NoError(t, err)

	// check if validator is jailed
	validator, err := stakingKeeper.GetValidator(ctx, addrVal1)
	require.NoError(t, err)
	require.True(t, validator.Jailed)
}

// TestICADelegateUndelegate tests that an ICA account can undelegate
// sequentially right after delegating.
func TestICADelegateUndelegate(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		accountKeeper = f.accountKeeper
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
	)

	// Create a delegator and validator (the delegator will be an ICA account)
	delegateAmount := math.NewInt(1000)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	require.NoError(t, err)
	delegateCoin := sdk.NewCoin(bondDenom, delegateAmount)
	icaAccountAddress := createICAAccount(ctx, accountKeeper)

	// Fund ICA account
	err = bankKeeper.MintCoins(ctx, minttypes.ModuleName, sdk.NewCoins(delegateCoin))
	require.NoError(t, err)
	err = bankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, icaAccountAddress, sdk.NewCoins(delegateCoin))
	require.NoError(t, err)

	addresses := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, math.NewInt(0))
	pubKeys := simtestutil.CreateTestPubKeys(1)
	validatorAddress := sdk.ValAddress(addresses[0])
	validator, err := types.NewValidator(validatorAddress.String(), pubKeys[0], types.Description{})
	require.NoError(t, err)

	validator.DelegatorShares = math.LegacyNewDec(1_000_000)
	validator.Tokens = math.NewInt(1_000_000)
	validator.LiquidShares = math.LegacyNewDec(0)
	stakingKeeper.SetValidator(ctx, validator)

	delegateMsg := types.MsgDelegate{
		DelegatorAddress: icaAccountAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	}

	undelegateMsg := types.MsgUndelegate{
		DelegatorAddress: icaAccountAddress.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegateCoin,
	}

	msgServer := keeper.NewMsgServerImpl(stakingKeeper)

	// Delegate normally
	_, err = msgServer.Delegate(ctx, &delegateMsg)
	require.NoError(t, err, "no error expected when delegating")

	// Confirm delegation record
	_, err = stakingKeeper.GetDelegation(ctx, icaAccountAddress, validatorAddress)
	require.NoError(t, err, "delegation should have been found")

	// Confirm liquid staking totals were incremented
	expectedTotalLiquidStaked := delegateAmount.Int64()
	actualTotalLiquidStaked := stakingKeeper.GetTotalLiquidStakedTokens(ctx).Int64()
	require.Equal(t, expectedTotalLiquidStaked, actualTotalLiquidStaked, "total liquid staked tokens after delegation")

	validator, err = stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err, "validator should have been found")
	require.Equal(t, math.LegacyNewDecFromInt(delegateAmount), validator.LiquidShares, "validator liquid shares after delegation")

	// Try to undelegate
	_, err = msgServer.Undelegate(ctx, &undelegateMsg)
	require.NoError(t, err, "no error expected when sequentially undelegating")

	// Confirm delegation record was removed
	_, err = stakingKeeper.GetDelegation(ctx, icaAccountAddress, validatorAddress)
	require.True(t, errors.Is(err, types.ErrNoDelegation), "delegation have been found")

	// Confirm liquid staking totals were decremented
	actualTotalLiquidStaked = stakingKeeper.GetTotalLiquidStakedTokens(ctx).Int64()
	require.Zero(t, actualTotalLiquidStaked, "total liquid staked tokens after undelegation")

	validator, err = stakingKeeper.GetValidator(ctx, validatorAddress)
	require.NoError(t, err, "validator should have been found")
	require.Equal(t, math.LegacyZeroDec(), validator.LiquidShares, "validator liquid shares after undelegation")
}

func TestTokenizeAndRedeemVestedDelegation(t *testing.T) {
	f := initFixture(t)

	ctx := f.sdkCtx
	var (
		stakingKeeper = f.stakingKeeper
		bankKeeper    = f.bankKeeper
		accountKeeper = f.accountKeeper
	)

	addrs := simtestutil.AddTestAddrs(bankKeeper, stakingKeeper, ctx, 1, stakingKeeper.TokensFromConsensusPower(ctx, 10000))
	addrAcc1 := addrs[0]
	addrVal1 := sdk.ValAddress(addrAcc1)

	// Original vesting mount (OV)
	originalVesting := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100_000)))
	startTime := time.Now()
	endTime := time.Now().Add(24 * time.Hour)

	// Create vesting account
	lastAccNum := uint64(1000)
	baseAcc := authtypes.NewBaseAccountWithAddress(addrAcc1)
	require.NoError(t, baseAcc.SetAccountNumber(atomic.AddUint64(&lastAccNum, 1)))

	continuousVestingAccount, err := vestingtypes.NewContinuousVestingAccount(
		baseAcc,
		originalVesting,
		startTime.Unix(),
		endTime.Unix(),
	)
	require.NoError(t, err)
	accountKeeper.SetAccount(ctx, continuousVestingAccount)

	pubKeys := simtestutil.CreateTestPubKeys(1)
	pk1 := pubKeys[0]

	// Create Validators and Delegation
	val1 := testutil.NewValidator(t, addrVal1, pk1)
	val1.Status = types.Bonded
	stakingKeeper.SetValidator(ctx, val1)
	stakingKeeper.SetValidatorByPowerIndex(ctx, val1)
	err = stakingKeeper.SetValidatorByConsAddr(ctx, val1)
	require.NoError(t, err)

	// Delegate all the vesting coins
	originalVestingAmount := originalVesting.AmountOf(sdk.DefaultBondDenom)
	err = delegateCoinsFromAccount(ctx, *stakingKeeper, addrAcc1, originalVestingAmount, val1)
	require.NoError(t, err)

	// Apply TM updates
	applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)

	_, err = stakingKeeper.GetDelegation(ctx, addrAcc1, addrVal1)
	require.NoError(t, err)

	// Check vesting account data
	// V=100, V'=0, DV=100, DF=0
	acc := accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, originalVesting, acc.GetVestingCoins(ctx.BlockTime()))
	require.Empty(t, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, originalVesting, acc.GetDelegatedVesting())
	require.Empty(t, acc.GetDelegatedFree())

	msgServer := keeper.NewMsgServerImpl(stakingKeeper)

	// Vest half the original vesting coins
	vestHalfTime := startTime.Add(time.Duration(float64(endTime.Sub(startTime).Nanoseconds()) / float64(2)))
	ctx = ctx.WithBlockTime(vestHalfTime)

	// expect that half of the orignal vesting coins are vested
	expVestedCoins := originalVesting.QuoInt(math.NewInt(2))

	// Check vesting account data
	// V=50, V'=50, DV=100, DF=0
	acc = accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, expVestedCoins, acc.GetVestingCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, originalVesting, acc.GetDelegatedVesting())
	require.Empty(t, acc.GetDelegatedFree())

	// Expect that tokenizing all the delegated coins fails
	// since only the half are vested
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrAcc1.String(),
		ValidatorAddress:    addrVal1.String(),
		Amount:              originalVesting[0],
		TokenizedShareOwner: addrAcc1.String(),
	})
	require.Error(t, err)

	// Tokenize the delegated vested coins
	_, err = msgServer.TokenizeShares(ctx, &types.MsgTokenizeShares{
		DelegatorAddress:    addrAcc1.String(),
		ValidatorAddress:    addrVal1.String(),
		Amount:              sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: originalVestingAmount.Quo(math.NewInt(2))},
		TokenizedShareOwner: addrAcc1.String(),
	})
	require.NoError(t, err)

	shareDenom := addrVal1.String() + "/1"

	// Redeem the tokens
	_, err = msgServer.RedeemTokensForShares(ctx,
		&types.MsgRedeemTokensForShares{
			DelegatorAddress: addrAcc1.String(),
			Amount:           sdk.Coin{Denom: shareDenom, Amount: originalVestingAmount.Quo(math.NewInt(2))},
		},
	)
	require.NoError(t, err)

	// After the redemption of the tokens, the vesting delegations should be evenly distributed
	// V=50, V'=50, DV=100, DF=50
	acc = accountKeeper.GetAccount(ctx, addrAcc1).(*vestingtypes.ContinuousVestingAccount)
	require.Equal(t, expVestedCoins, acc.GetVestingCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetVestedCoins(ctx.BlockTime()))
	require.Equal(t, expVestedCoins, acc.GetDelegatedVesting())
	require.Equal(t, expVestedCoins, acc.GetDelegatedFree())
}

// Helper function to create 32-length account
// Used to mock an liquid staking provider's ICA account
func createICAAccount(ctx sdk.Context, ak accountkeeper.AccountKeeper) sdk.AccAddress {
	icahost := "icahost"
	connectionID := "connection-0"
	portID := icahost

	moduleAddress := authtypes.NewModuleAddress(icahost)
	icaAddress := sdk.AccAddress(sdkaddress.Derive(moduleAddress, []byte(connectionID+portID)))

	account := ak.NewAccountWithAddress(ctx, icaAddress)
	ak.SetAccount(ctx, account)

	return icaAddress
}
