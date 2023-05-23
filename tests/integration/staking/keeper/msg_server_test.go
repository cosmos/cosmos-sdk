package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"gotest.tools/v3/assert"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
)

func TestCancelUnbondingDelegation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	bondDenom := f.stakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := f.stakingKeeper.GetNotBondedPool(ctx)
	startTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, 5)

	assert.NilError(t, testutil.FundModuleAccount(ctx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(f.stakingKeeper.BondDenom(ctx), startTokens))))
	f.accountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := f.bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), f.stakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, ctx, 2, sdk.NewInt(10000))
	valAddr := sdk.ValAddress(addrs[0])
	delegatorAddr := addrs[1]

	// setup a new validator with bonded status
	validator, err := types.NewValidator(valAddr, PKs[0], types.NewDescription("Validator", "", "", "", ""))
	validator.Status = types.Bonded
	assert.NilError(t, err)
	f.stakingKeeper.SetValidator(ctx, validator)

	validatorAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	assert.NilError(t, err)

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(f.stakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
		0,
	)

	// set and retrieve a record
	f.stakingKeeper.SetUnbondingDelegation(ctx, ubd)
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
				Amount:           sdk.NewCoin(f.stakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
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
				Amount:           sdk.NewCoin(f.stakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
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
				Amount:           sdk.NewCoin("dump_coin", sdk.NewInt(4)),
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

func TestRotateConsPubKey(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	stakingKeeper := f.stakingKeeper
	bankKeeper := f.bankKeeper
	accountKeeper := f.accountKeeper

	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	bondDenom := stakingKeeper.BondDenom(ctx)

	params := stakingKeeper.GetParams(ctx)
	params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
	err := stakingKeeper.SetParams(ctx, params)
	assert.NilError(t, err)

	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 5, stakingKeeper.TokensFromConsensusPower(ctx, 100))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	// create 5 validators
	for i := 0; i < 5; i++ {
		comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))

		msg, err := types.NewMsgCreateValidator(valAddrs[i], PKs[i], sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 30)),
			types.Description{Moniker: "NewVal"}, comm, math.OneInt())
		require.NoError(t, err)
		_, err = msgServer.CreateValidator(ctx, msg)
		require.NoError(t, err)
	}

	// call endblocker to update the validator state
	_, err = stakingKeeper.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeader().Height + 1))
	require.NoError(t, err)

	keyRotationFee := stakingKeeper.KeyRotationFee(ctx)

	validators := stakingKeeper.GetAllValidators(ctx)
	require.GreaterOrEqual(t, len(validators), 5)

	testCases := []struct {
		name           string
		malleate       func() sdk.Context
		pass           bool
		validator      sdk.ValAddress
		newPubKey      cryptotypes.PubKey
		expErrMsg      string
		expHistoryObjs int
		fees           sdk.Coin
	}{
		{
			name: "successful consensus pubkey rotation",
			malleate: func() sdk.Context {
				return ctx
			},
			validator:      validators[0].GetOperator(),
			newPubKey:      PKs[499],
			pass:           true,
			expHistoryObjs: 1,
			fees:           keyRotationFee,
		},
		{
			name: "non existing validator check",
			malleate: func() sdk.Context {
				return ctx
			},
			validator: sdk.ValAddress("non_existing_val"),
			newPubKey: PKs[498],
			pass:      false,
			expErrMsg: "validator does not exist",
		},
		{
			name: "pubkey already associated with another validator",
			malleate: func() sdk.Context {
				return ctx
			},
			validator: validators[0].GetOperator(),
			newPubKey: validators[1].ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey),
			pass:      false,
			expErrMsg: "consensus pubkey is already used for a validator",
		},
		{
			name: "consensus pubkey rotation limit check",
			malleate: func() sdk.Context {
				params := stakingKeeper.GetParams(ctx)
				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				err := stakingKeeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg, err := types.NewMsgRotateConsPubKey(
					validators[1].GetOperator(),
					PKs[498],
				)
				require.NoError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				require.NoError(t, err)

				return ctx
			},
			validator: validators[1].GetOperator(),
			newPubKey: PKs[497],
			pass:      false,
			expErrMsg: "exceeding maximum consensus pubkey rotations within unbonding period",
		},
		{
			name: "limit reached, but should rotate after the unbonding period",
			malleate: func() sdk.Context {
				params := stakingKeeper.GetParams(ctx)
				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				err := stakingKeeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg, err := types.NewMsgRotateConsPubKey(
					validators[3].GetOperator(),
					PKs[495],
				)

				require.NoError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

				// this shouldn't remove the existing keys from waiting queue since unbonding time isn't reached
				_, err = stakingKeeper.EndBlocker(ctx)
				require.NoError(t, err)
				// stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(ctx, ctx.BlockHeader().Time)

				msg, err = types.NewMsgRotateConsPubKey(
					validators[3].GetOperator(),
					PKs[494],
				)

				require.NoError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				require.Error(t, err)

				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

				newCtx := ctx.WithBlockTime(ctx.BlockHeader().Time.Add(stakingKeeper.UnbondingTime(ctx)))
				newCtx = newCtx.WithBlockHeight(newCtx.BlockHeight() + 1)
				// this should remove keys from waiting queue since unbonding time is reached
				_, err = stakingKeeper.EndBlocker(newCtx)
				require.NoError(t, err)
				// stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(newCtx, newCtx.BlockHeader().Time)

				return newCtx
			},
			validator:      validators[3].GetOperator(),
			newPubKey:      PKs[494],
			pass:           true,
			expErrMsg:      "",
			expHistoryObjs: 2,
			fees:           keyRotationFee,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			newCtx := testCase.malleate()
			oldDistrBalance := bankKeeper.GetBalance(newCtx, accountKeeper.GetModuleAddress(distrtypes.ModuleName), bondDenom)
			msg, err := types.NewMsgRotateConsPubKey(
				testCase.validator,
				testCase.newPubKey,
			)
			require.NoError(t, err)

			_, err = msgServer.RotateConsPubKey(newCtx, msg)

			if testCase.pass {
				require.NoError(t, err)

				_, err = stakingKeeper.EndBlocker(newCtx)
				require.NoError(t, err)

				// rotation fee payment from sender to distrtypes
				newDistrBalance := bankKeeper.GetBalance(newCtx, accountKeeper.GetModuleAddress(distrtypes.ModuleName), bondDenom)
				require.Equal(t, newDistrBalance, oldDistrBalance.Add(testCase.fees))

				// validator consensus pubkey update check
				validator, found := stakingKeeper.GetValidator(newCtx, testCase.validator)
				require.True(t, found)

				consAddr, err := validator.GetConsAddr()
				require.NoError(t, err)
				require.Equal(t, consAddr.String(), sdk.ConsAddress(testCase.newPubKey.Address()).String())

				// consensus rotation history set check
				historyObjects := stakingKeeper.GetValidatorConsPubKeyRotationHistory(newCtx, testCase.validator)
				require.Len(t, historyObjects, testCase.expHistoryObjs)
				historyObjects = stakingKeeper.GetBlockConsPubKeyRotationHistory(newCtx, newCtx.BlockHeight())
				require.Len(t, historyObjects, 1)
			} else {
				require.Error(t, err)
				require.Equal(t, err.Error(), testCase.expErrMsg)
			}
		})
	}
}
