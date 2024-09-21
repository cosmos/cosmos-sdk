package keeper_test

import (
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	"cosmossdk.io/x/bank/testutil"
	pooltypes "cosmossdk.io/x/protocolpool/types"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	validator, err := types.NewValidator(valAddr.String(), PKs[0], types.NewDescription("Validator", "", "", "", "", types.Metadata{}))
	validator.Status = types.Bonded
	assert.NilError(t, err)
	assert.NilError(t, f.stakingKeeper.SetValidator(ctx, validator))

	validatorAddr, err := sdk.ValAddressFromBech32(validator.OperatorAddress)
	assert.NilError(t, err)

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(bondDenom, 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.HeaderInfo().Time.Add(time.Minute*10),
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

func TestRotateConsPubKey(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	ctx := f.sdkCtx
	stakingKeeper := f.stakingKeeper
	bankKeeper := f.bankKeeper
	accountKeeper := f.accountKeeper

	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	bondDenom, err := stakingKeeper.BondDenom(ctx)
	assert.NilError(t, err)

	params, err := stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)

	params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
	err = stakingKeeper.Params.Set(ctx, params)
	assert.NilError(t, err)

	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 5, stakingKeeper.TokensFromConsensusPower(ctx, 100))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	// create 5 validators
	for i := 0; i < 5; i++ {
		comm := types.NewCommissionRates(math.LegacyNewDec(0), math.LegacyNewDec(0), math.LegacyNewDec(0))
		acc := f.accountKeeper.NewAccountWithAddress(ctx, sdk.AccAddress(valAddrs[i]))
		f.accountKeeper.SetAccount(ctx, acc)
		msg, err := types.NewMsgCreateValidator(valAddrs[i].String(), PKs[i], sdk.NewCoin(sdk.DefaultBondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 30)),
			types.Description{Moniker: "NewVal"}, comm, math.OneInt())
		assert.NilError(t, err)
		_, err = msgServer.CreateValidator(ctx, msg)
		assert.NilError(t, err)
	}

	// call endblocker to update the validator state
	_, err = stakingKeeper.EndBlocker(ctx.WithBlockHeight(ctx.BlockHeader().Height + 1))
	assert.NilError(t, err)

	params, err = stakingKeeper.Params.Get(ctx)
	assert.NilError(t, err)

	validators, err := stakingKeeper.GetAllValidators(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(validators) >= 5, true)

	testCases := []struct {
		name           string
		malleate       func() sdk.Context
		pass           bool
		validator      string
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
			fees:           params.KeyRotationFee,
		},
		{
			name: "non existing validator check",
			malleate: func() sdk.Context {
				return ctx
			},
			validator: sdk.ValAddress("non_existing_val").String(),
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
			expErrMsg: "validator already exist for this pubkey; must use new validator pubkey",
		},
		{
			name: "consensus pubkey rotation limit check",
			malleate: func() sdk.Context {
				params, err := stakingKeeper.Params.Get(ctx)
				assert.NilError(t, err)

				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				err = stakingKeeper.Params.Set(ctx, params)
				assert.NilError(t, err)

				msg, err := types.NewMsgRotateConsPubKey(
					validators[1].GetOperator(),
					PKs[498],
				)
				assert.NilError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				assert.NilError(t, err)

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
				params, err := stakingKeeper.Params.Get(ctx)
				assert.NilError(t, err)

				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				err = stakingKeeper.Params.Set(ctx, params)
				assert.NilError(t, err)

				msg, err := types.NewMsgRotateConsPubKey(
					validators[3].GetOperator(),
					PKs[495],
				)

				assert.NilError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				assert.NilError(t, err)
				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

				// this shouldn't remove the existing keys from waiting queue since unbonding time isn't reached
				_, err = stakingKeeper.EndBlocker(ctx)
				assert.NilError(t, err)

				msg, err = types.NewMsgRotateConsPubKey(
					validators[3].GetOperator(),
					PKs[494],
				)

				assert.NilError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				assert.Error(t, err, "exceeding maximum consensus pubkey rotations within unbonding period")

				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

				newCtx := ctx.WithHeaderInfo(header.Info{Height: ctx.BlockHeight() + 1, Time: ctx.HeaderInfo().Time.Add(params.UnbondingTime)}).WithBlockHeight(ctx.BlockHeight() + 1)
				// this should remove keys from waiting queue since unbonding time is reached
				_, err = stakingKeeper.EndBlocker(newCtx)
				assert.NilError(t, err)

				return newCtx
			},
			validator:      validators[3].GetOperator(),
			newPubKey:      PKs[494],
			pass:           true,
			expErrMsg:      "",
			expHistoryObjs: 2,
			fees:           params.KeyRotationFee,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			newCtx := testCase.malleate()
			oldDistrBalance := bankKeeper.GetBalance(newCtx, accountKeeper.GetModuleAddress(pooltypes.ModuleName), bondDenom)
			msg, err := types.NewMsgRotateConsPubKey(
				testCase.validator,
				testCase.newPubKey,
			)
			assert.NilError(t, err)

			_, err = msgServer.RotateConsPubKey(newCtx, msg)

			if testCase.pass {
				assert.NilError(t, err)

				_, err = stakingKeeper.EndBlocker(newCtx)
				assert.NilError(t, err)

				// rotation fee payment from sender to distrtypes
				newDistrBalance := bankKeeper.GetBalance(newCtx, accountKeeper.GetModuleAddress(pooltypes.ModuleName), bondDenom)
				assert.DeepEqual(t, newDistrBalance, oldDistrBalance.Add(testCase.fees))

				valBytes, err := stakingKeeper.ValidatorAddressCodec().StringToBytes(testCase.validator)
				assert.NilError(t, err)

				// validator consensus pubkey update check
				validator, err := stakingKeeper.GetValidator(newCtx, valBytes)
				assert.NilError(t, err)

				consAddr, err := validator.GetConsAddr()
				assert.NilError(t, err)
				assert.DeepEqual(t, consAddr, testCase.newPubKey.Address().Bytes())

				// consensus rotation history set check
				historyObjects, err := stakingKeeper.GetValidatorConsPubKeyRotationHistory(newCtx, valBytes)
				assert.NilError(t, err)
				assert.Equal(t, len(historyObjects), testCase.expHistoryObjs)

				historyObjects, err = stakingKeeper.GetBlockConsPubKeyRotationHistory(newCtx)
				assert.NilError(t, err)
				assert.Equal(t, len(historyObjects), 1)

			} else {
				assert.ErrorContains(t, err, testCase.expErrMsg)
			}
		})
	}
}
