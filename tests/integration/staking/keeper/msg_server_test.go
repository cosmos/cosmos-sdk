package keeper_test

import (
	"math/big"
	"testing"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestCancelUnbondingDelegation(t *testing.T) {
	// setup the app
	var (
		stakingKeeper *keeper.Keeper
		bankKeeper    bankkeeper.Keeper
		accountKeeper authkeeper.AccountKeeper
	)
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
		&stakingKeeper, &bankKeeper, &accountKeeper)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	bondDenom := stakingKeeper.BondDenom(ctx)

	// set the not bonded pool module account
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)
	startTokens := stakingKeeper.TokensFromConsensusPower(ctx, 5)

	assert.NilError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), startTokens))))
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	moduleBalance := bankKeeper.GetBalance(ctx, notBondedPool.GetAddress(), stakingKeeper.BondDenom(ctx))
	assert.DeepEqual(t, sdk.NewInt64Coin(bondDenom, startTokens.Int64()), moduleBalance)

	// accounts
	delAddrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 2, sdk.NewInt(10000))
	validators := stakingKeeper.GetValidators(ctx, 10)
	assert.Equal(t, len(validators), 1)

	validatorAddr, err := sdk.ValAddressFromBech32(validators[0].OperatorAddress)
	assert.NilError(t, err)
	delegatorAddr := delAddrs[0]

	// setting the ubd entry
	unbondingAmount := sdk.NewInt64Coin(stakingKeeper.BondDenom(ctx), 5)
	ubd := types.NewUnbondingDelegation(
		delegatorAddr, validatorAddr, 10,
		ctx.BlockTime().Add(time.Minute*10),
		unbondingAmount.Amount,
		0,
	)

	// set and retrieve a record
	stakingKeeper.SetUnbondingDelegation(ctx, ubd)
	resUnbond, found := stakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	assert.Assert(t, found)
	assert.DeepEqual(t, ubd, resUnbond)

	testCases := []struct {
		Name      string
		ExceptErr bool
		req       types.MsgCancelUnbondingDelegation
		expErrMsg string
	}{
		{
			Name:      "invalid height",
			ExceptErr: true,
			req: types.MsgCancelUnbondingDelegation{
				DelegatorAddress: resUnbond.DelegatorAddress,
				ValidatorAddress: resUnbond.ValidatorAddress,
				Amount:           sdk.NewCoin(stakingKeeper.BondDenom(ctx), sdk.NewInt(4)),
				CreationHeight:   0,
			},
			expErrMsg: "unbonding delegation entry is not found at block height",
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
			expErrMsg: "invalid coin denomination",
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
			expErrMsg: "validator does not exist",
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
			expErrMsg: "decoding bech32 failed",
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
			expErrMsg: "amount is greater than the unbonding delegation entry balance",
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
				assert.ErrorContains(t, err, testCase.expErrMsg)
			} else {
				assert.NilError(t, err)
				balanceForNotBondedPool := bankKeeper.GetBalance(ctx, sdk.AccAddress(notBondedPool.GetAddress()), bondDenom)
				assert.DeepEqual(t, balanceForNotBondedPool, moduleBalance.Sub(testCase.req.Amount))
				moduleBalance = moduleBalance.Sub(testCase.req.Amount)
			}
		})
	}
}

func TestRotateConsPubKey(t *testing.T) {
	// setup the app
	var (
		stakingKeeper *keeper.Keeper
		bankKeeper    bankkeeper.Keeper
		accountKeeper authkeeper.AccountKeeper
	)
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
		&accountKeeper, &bankKeeper, &stakingKeeper)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	msgServer := keeper.NewMsgServerImpl(stakingKeeper)
	bondDenom := stakingKeeper.BondDenom(ctx)

	params := stakingKeeper.GetParams(ctx)
	params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
	params.MaxConsPubkeyRotations = types.DefaultMaxConsPubKeyRotations
	err = stakingKeeper.SetParams(ctx, params)
	assert.NilError(t, err)

	addrs := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 5, stakingKeeper.TokensFromConsensusPower(ctx, 1000000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrs)

	validators := stakingKeeper.GetAllValidators(ctx)
	require.Len(t, validators, 1)

	// create 5 validators
	for i := 0; i < 5; i++ {
		val := testutil.NewValidator(t, valAddrs[i], PKs[i])
		stakingKeeper.SetValidator(ctx, val)
		stakingKeeper.SetValidatorByConsAddr(ctx, val)
		stakingKeeper.SetNewValidatorByPowerIndex(ctx, val)
	}

	keyRotationFee := stakingKeeper.KeyRotationFee(ctx)

	validators = stakingKeeper.GetAllValidators(ctx)
	require.GreaterOrEqual(t, len(validators), 5)
	validators = validators[1:]

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
				params.MaxConsPubkeyRotations = 1
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
			name: "two rotations within unbonding period",
			malleate: func() sdk.Context {
				params := stakingKeeper.GetParams(ctx)
				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				params.MaxConsPubkeyRotations = 2
				err := stakingKeeper.SetParams(ctx, params)
				require.NoError(t, err)

				msg, err := types.NewMsgRotateConsPubKey(
					validators[2].GetOperator(),
					PKs[497],
				)
				require.NoError(t, err)
				_, err = msgServer.RotateConsPubKey(ctx, msg)
				require.NoError(t, err)
				ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

				return ctx
			},
			validator:      validators[2].GetOperator(),
			newPubKey:      PKs[496],
			pass:           true,
			fees:           calculateFee(keyRotationFee, 1),
			expHistoryObjs: 2,
		},
		{
			name: "limit reached, but should rotate after the unbonding period",
			malleate: func() sdk.Context {
				params := stakingKeeper.GetParams(ctx)
				params.KeyRotationFee = sdk.NewInt64Coin(bondDenom, 10)
				params.MaxConsPubkeyRotations = 1
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
				stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(ctx, ctx.BlockHeader().Time)

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
				stakingKeeper.UpdateAllMaturedConsKeyRotatedKeys(newCtx, newCtx.BlockHeader().Time)

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

func calculateFee(fee sdk.Coin, rotationsMade int64) sdk.Coin {
	fees := sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(rotationsMade)), nil))
	fees = fee.Amount.Mul(fees)
	return sdk.NewCoin(fee.Denom, fees)
}
