package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// SimulateMsgCreateValidator
func SimulateMsgCreateValidator(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		description := stake.Description{
			Moniker: simulation.RandStringOfLength(r, 10),
		}
		key := simulation.RandomKey(r, keys)
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		amount := m.GetAccount(ctx, address).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simulation.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "no-operation", nil
		}
		msg := stake.MsgCreateValidator{
			Description:   description,
			ValidatorAddr: address,
			DelegatorAddr: address,
			PubKey:        pubkey,
			Delegation:    sdk.NewCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCreateValidator/%v", result.IsOK()))
		// require.True(t, result.IsOK(), "expected OK result but instead got %v", result)
		action = fmt.Sprintf("TestMsgCreateValidator: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgEditValidator
func SimulateMsgEditValidator(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		description := stake.Description{
			Moniker:  simulation.RandStringOfLength(r, 10),
			Identity: simulation.RandStringOfLength(r, 10),
			Website:  simulation.RandStringOfLength(r, 10),
			Details:  simulation.RandStringOfLength(r, 10),
		}
		key := simulation.RandomKey(r, keys)
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		msg := stake.MsgEditValidator{
			Description:   description,
			ValidatorAddr: address,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgEditValidator/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgEditValidator: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgDelegate
func SimulateMsgDelegate(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := simulation.RandomKey(r, keys)
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := simulation.RandomKey(r, keys)
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simulation.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "no-operation", nil
		}
		msg := stake.MsgDelegate{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
			Delegation:    sdk.NewCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgDelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgDelegate: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginUnbonding
func SimulateMsgBeginUnbonding(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := simulation.RandomKey(r, keys)
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := simulation.RandomKey(r, keys)
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simulation.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "no-operation", nil
		}
		msg := stake.MsgBeginUnbonding{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
			SharesAmount:  sdk.NewRatFromInt(amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgBeginUnbonding/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgBeginUnbonding: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteUnbonding
func SimulateMsgCompleteUnbonding(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		validatorKey := simulation.RandomKey(r, keys)
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := simulation.RandomKey(r, keys)
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		msg := stake.MsgCompleteUnbonding{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCompleteUnbonding/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgCompleteUnbonding: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginRedelegate
func SimulateMsgBeginRedelegate(m auth.AccountMapper, k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		sourceValidatorKey := simulation.RandomKey(r, keys)
		sourceValidatorAddress := sdk.AccAddress(sourceValidatorKey.PubKey().Address())
		destValidatorKey := simulation.RandomKey(r, keys)
		destValidatorAddress := sdk.AccAddress(destValidatorKey.PubKey().Address())
		delegatorKey := simulation.RandomKey(r, keys)
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		// TODO
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = simulation.RandomAmount(r, amount)
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "no-operation", nil
		}
		msg := stake.MsgBeginRedelegate{
			DelegatorAddr:    delegatorAddress,
			ValidatorSrcAddr: sourceValidatorAddress,
			ValidatorDstAddr: destValidatorAddress,
			SharesAmount:     sdk.NewRatFromInt(amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgBeginRedelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgBeginRedelegate: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteRedelegate
func SimulateMsgCompleteRedelegate(k stake.Keeper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
		validatorSrcKey := simulation.RandomKey(r, keys)
		validatorSrcAddress := sdk.AccAddress(validatorSrcKey.PubKey().Address())
		validatorDstKey := simulation.RandomKey(r, keys)
		validatorDstAddress := sdk.AccAddress(validatorDstKey.PubKey().Address())
		delegatorKey := simulation.RandomKey(r, keys)
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		msg := stake.MsgCompleteRedelegate{
			DelegatorAddr:    delegatorAddress,
			ValidatorSrcAddr: validatorSrcAddress,
			ValidatorDstAddr: validatorDstAddress,
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		event(fmt.Sprintf("stake/MsgCompleteRedelegate/%v", result.IsOK()))
		action = fmt.Sprintf("TestMsgCompleteRedelegate: ok %v, msg %s", result.IsOK(), msg.GetSignBytes())
		return action, nil
	}
}

// Setup
func Setup(mapp *mock.App, k stake.Keeper) simulation.RandSetup {
	return func(r *rand.Rand, privKeys []crypto.PrivKey) {
		ctx := mapp.NewContext(false, abci.Header{})
		gen := stake.DefaultGenesisState()
		gen.Params.InflationMax = sdk.NewRat(0)
		gen.Params.InflationMin = sdk.NewRat(0)
		stake.InitGenesis(ctx, k, gen)
		params := k.GetParams(ctx)
		denom := params.BondDenom
		loose := sdk.ZeroInt()
		mapp.AccountMapper.IterateAccounts(ctx, func(acc auth.Account) bool {
			balance := simulation.RandomAmount(r, sdk.NewInt(1000000))
			acc.SetCoins(acc.GetCoins().Plus(sdk.Coins{sdk.NewCoin(denom, balance)}))
			mapp.AccountMapper.SetAccount(ctx, acc)
			loose = loose.Add(balance)
			return false
		})
		pool := k.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Add(sdk.NewRat(loose.Int64(), 1))
		k.SetPool(ctx, pool)
	}
}
