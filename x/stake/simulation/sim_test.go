package simulation

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/stake"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

var (
	stats = make(map[string]int)
)

// ModuleInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func ModuleInvariants(ck bank.Keeper, k stake.Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		SupplyInvariants(ck, k)(t, app, log)
		PositivePowerInvariant(k)(t, app, log)
		ValidatorSetInvariant(k)(t, app, log)
	}
}

// SupplyInvariants checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
func SupplyInvariants(ck bank.Keeper, k stake.Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		ctx := app.NewContext(false, abci.Header{})
		pool := k.GetPool(ctx)

		// Loose tokens should equal coin supply
		loose := sdk.ZeroInt()
		app.AccountMapper.IterateAccounts(ctx, func(acc auth.Account) bool {
			loose = loose.Add(acc.GetCoins().AmountOf("steak"))
			return false
		})
		require.True(t, pool.LooseTokens.RoundInt64() == loose.Int64(), "expected loose tokens to equal total steak held by accounts - pool.LooseTokens: %v, sum of account tokens: %v\nlog: %s",
			pool.LooseTokens, loose, log)
		stats["stake/invariant/looseTokens"] += 1

		// Bonded tokens should equal sum of tokens with bonded validators
		bonded := sdk.ZeroRat()
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetPower())
			}
			return false
		})
		require.True(t, pool.BondedTokens.Equal(bonded), "expected bonded tokens to equal total steak held by bonded validators\nlog: %s", log)
		stats["stake/invariant/bondedTokens"] += 1

		// TODO Inflation check on total supply
	}
}

// PositivePowerInvariant checks that all stored validators have > 0 power
func PositivePowerInvariant(k stake.Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		ctx := app.NewContext(false, abci.Header{})
		k.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) bool {
			require.True(t, validator.GetPower().GT(sdk.ZeroRat()), "validator with non-positive power stored")
			return false
		})
		stats["stake/invariant/positivePower"] += 1
	}
}

// ValidatorSetInvariant checks equivalence of Tendermint validator set and SDK validator set
func ValidatorSetInvariant(k stake.Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		// TODO
	}
}

// SimulateMsgCreateValidator
func SimulateMsgCreateValidator(m auth.AccountMapper, k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		description := stake.Description{
			Moniker: mock.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		amount := m.GetAccount(ctx, address).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgCreateValidator{
			Description:   description,
			ValidatorAddr: address,
			DelegatorAddr: address,
			PubKey:        pubkey,
			Delegation:    sdk.NewIntCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		stats[fmt.Sprintf("stake/createvalidator/%v", result.IsOK())] += 1
		// require.True(t, result.IsOK(), "expected OK result but instead got %v", result)
		action = fmt.Sprintf("TestMsgCreateValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgEditValidator
func SimulateMsgEditValidator(k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		description := stake.Description{
			Moniker:  mock.RandStringOfLength(r, 10),
			Identity: mock.RandStringOfLength(r, 10),
			Website:  mock.RandStringOfLength(r, 10),
			Details:  mock.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
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
		stats[fmt.Sprintf("stake/editvalidator/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgEditValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgDelegate
func SimulateMsgDelegate(m auth.AccountMapper, k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
		}
		msg := stake.MsgDelegate{
			DelegatorAddr: delegatorAddress,
			ValidatorAddr: validatorAddress,
			Delegation:    sdk.NewIntCoin(denom, amount),
		}
		require.Nil(t, msg.ValidateBasic(), "expected msg to pass ValidateBasic: %s", msg.GetSignBytes())
		ctx, write := ctx.CacheContext()
		result := stake.NewHandler(k)(ctx, msg)
		if result.IsOK() {
			write()
		}
		stats[fmt.Sprintf("stake/delegate/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgDelegate: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginUnbonding
func SimulateMsgBeginUnbonding(m auth.AccountMapper, k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
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
		stats[fmt.Sprintf("stake/beginunbonding/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgBeginUnbonding: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteUnbonding
func SimulateMsgCompleteUnbonding(k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		validatorKey := keys[r.Intn(len(keys))]
		validatorAddress := sdk.AccAddress(validatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
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
		stats[fmt.Sprintf("stake/completeunbonding/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgCompleteUnbonding with %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgBeginRedelegate
func SimulateMsgBeginRedelegate(m auth.AccountMapper, k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		sourceValidatorKey := keys[r.Intn(len(keys))]
		sourceValidatorAddress := sdk.AccAddress(sourceValidatorKey.PubKey().Address())
		destValidatorKey := keys[r.Intn(len(keys))]
		destValidatorAddress := sdk.AccAddress(destValidatorKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
		delegatorAddress := sdk.AccAddress(delegatorKey.PubKey().Address())
		// TODO
		amount := m.GetAccount(ctx, delegatorAddress).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		if amount.Equal(sdk.ZeroInt()) {
			return "nop", nil
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
		stats[fmt.Sprintf("stake/beginredelegate/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgBeginRedelegate: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgCompleteRedelegate
func SimulateMsgCompleteRedelegate(k stake.Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		validatorSrcKey := keys[r.Intn(len(keys))]
		validatorSrcAddress := sdk.AccAddress(validatorSrcKey.PubKey().Address())
		validatorDstKey := keys[r.Intn(len(keys))]
		validatorDstAddress := sdk.AccAddress(validatorDstKey.PubKey().Address())
		delegatorKey := keys[r.Intn(len(keys))]
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
		stats[fmt.Sprintf("stake/completeredelegate/%v", result.IsOK())] += 1
		action = fmt.Sprintf("TestMsgCompleteRedelegate with %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulationSetup
func SimulationSetup(mapp *mock.App, k stake.Keeper) mock.RandSetup {
	return func(r *rand.Rand, privKeys []crypto.PrivKey) {
		ctx := mapp.NewContext(false, abci.Header{})
		stake.InitGenesis(ctx, k, stake.DefaultGenesisState())
		params := k.GetParams(ctx)
		denom := params.BondDenom
		loose := sdk.ZeroInt()
		mapp.AccountMapper.IterateAccounts(ctx, func(acc auth.Account) bool {
			balance := sdk.NewInt(int64(r.Intn(1000000)))
			acc.SetCoins(acc.GetCoins().Plus(sdk.Coins{sdk.NewIntCoin(denom, balance)}))
			mapp.AccountMapper.SetAccount(ctx, acc)
			loose = loose.Add(balance)
			return false
		})
		pool := k.GetPool(ctx)
		pool.LooseTokens = pool.LooseTokens.Add(sdk.NewRat(loose.Int64(), 1))
		k.SetPool(ctx, pool)
	}
}

// TestStakeWithRandomMessages
func TestStakeWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterWire(mapp.Cdc)
	mapper := mapp.AccountMapper
	coinKeeper := bank.NewKeeper(mapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeKeeper := stake.NewKeeper(mapp.Cdc, stakeKey, coinKeeper, stake.DefaultCodespace)
	mapp.Router().AddRoute("stake", stake.NewHandler(stakeKeeper))
	mapp.SetEndBlocker(func(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
		validatorUpdates := stake.EndBlocker(ctx, stakeKeeper)
		return abci.ResponseEndBlock{
			ValidatorUpdates: validatorUpdates,
		}
	})

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{stakeKey})
	if err != nil {
		panic(err)
	}

	mapp.SimpleRandomizedTestingFromSeed(
		t, 20, []mock.TestAndRunMsg{
			SimulateMsgCreateValidator(mapper, stakeKeeper),
			SimulateMsgEditValidator(stakeKeeper),
			SimulateMsgDelegate(mapper, stakeKeeper),
			// XXX TODO
			// SimulateMsgBeginUnbonding(mapper, stakeKeeper),
			SimulateMsgCompleteUnbonding(stakeKeeper),
			SimulateMsgBeginRedelegate(mapper, stakeKeeper),
			SimulateMsgCompleteRedelegate(stakeKeeper),
		}, []mock.RandSetup{
			SimulationSetup(mapp, stakeKeeper),
		}, []mock.Invariant{
			ModuleInvariants(coinKeeper, stakeKeeper),
		}, 10, 100, 500,
	)

	fmt.Printf("Stats: %v\n", stats)
}
