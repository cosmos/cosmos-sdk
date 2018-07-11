package stake

import (
	// "errors"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// ModuleInvariants runs all invariants of the stake module.
// Currently: total supply, positive power
func ModuleInvariants(ck bank.Keeper, k Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		SupplyInvariants(ck, k)(t, app, log)
		PositivePowerInvariant(k)(t, app, log)
		ValidatorSetInvariant(k)(t, app, log)
	}
}

// SupplyInvariants checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
func SupplyInvariants(ck bank.Keeper, k Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		ctx := app.NewContext(false, abci.Header{})
		pool := k.GetPool(ctx)

		// Loose tokens should equal coin supply
		loose := sdk.ZeroInt()
		app.AccountMapper.IterateAccounts(ctx, func(acc auth.Account) bool {
			loose = loose.Add(acc.GetCoins().AmountOf("steak"))
			return false
		})
		require.True(t, sdk.NewInt(pool.LooseTokens).Equal(loose), "expected loose tokens to equal total steak held by accounts")

		// Bonded tokens should equal sum of tokens with bonded validators
		// Unbonded tokens should equal sum of tokens with unbonded validators
		bonded := sdk.ZeroRat()
		unbonded := sdk.ZeroRat()
		k.IterateValidators(ctx, func(_ int64, validator sdk.Validator) bool {
			switch validator.GetStatus() {
			case sdk.Bonded:
				bonded = bonded.Add(validator.GetPower())
			case sdk.Unbonding:
				// TODO
			case sdk.Unbonded:
				unbonded = unbonded.Add(validator.GetPower())
			}
			return false
		})
		require.True(t, sdk.NewRat(pool.BondedTokens).Equal(bonded), "expected bonded tokens to equal total steak held by bonded validators")
		require.True(t, sdk.NewRat(pool.UnbondedTokens).Equal(unbonded), "expected unbonded tokens to equal total steak held by unbonded validators")

		// TODO Unbonding tokens

		// TODO Inflation check on total supply
	}
}

// PositivePowerInvariant checks that all stored validators have > 0 power
func PositivePowerInvariant(k Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		ctx := app.NewContext(false, abci.Header{})
		k.IterateValidatorsBonded(ctx, func(_ int64, validator sdk.Validator) bool {
			require.True(t, validator.GetPower().GT(sdk.ZeroRat()), "validator with non-positive power stored")
			return false
		})
	}
}

// ValidatorSetInvariant checks equivalence of Tendermint validator set and SDK validator set
func ValidatorSetInvariant(k Keeper) mock.Invariant {
	return func(t *testing.T, app *mock.App, log string) {
		// TODO
	}
}

// SimulateMsgCreateValidator
func SimulateMsgCreateValidator(m auth.AccountMapper, k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		denom := k.GetParams(ctx).BondDenom
		description := Description{
			Moniker: mock.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		amount := m.GetAccount(ctx, address).GetCoins().AmountOf(denom)
		if amount.GT(sdk.ZeroInt()) {
			amount = sdk.NewInt(int64(r.Intn(int(amount.Int64()))))
		}
		msg := MsgCreateValidator{
			Description:    description,
			ValidatorAddr:  address,
			PubKey:         pubkey,
			SelfDelegation: sdk.NewIntCoin(denom, amount),
		}
		action = fmt.Sprintf("TestMsgCreateValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgEditValidator
func SimulateMsgEditValidator(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		description := Description{
			Moniker:  mock.RandStringOfLength(r, 10),
			Identity: mock.RandStringOfLength(r, 10),
			Website:  mock.RandStringOfLength(r, 10),
			Details:  mock.RandStringOfLength(r, 10),
		}
		key := keys[r.Intn(len(keys))]
		pubkey := key.PubKey()
		address := sdk.AccAddress(pubkey.Address())
		msg := MsgEditValidator{
			Description:   description,
			ValidatorAddr: address,
		}
		action = fmt.Sprintf("TestMsgEditValidator: %s", msg.GetSignBytes())
		return action, nil
	}
}

// SimulateMsgDelegate
func SimulateMsgDelegate(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgDelegate with %s", "ok")
		return msg, nil
	}
}

// SimulateMsgBeginUnbonding
func SimulateMsgBeginUnbonding(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgBeginUnbonding with %s", "ok")
		return msg, nil
	}
}

// SimulateMsgCompleteUnbonding
func SimulateMsgCompleteUnbonding(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgCompleteUnbonding with %s", "ok")
		return msg, nil
	}
}

// SimulateMsgBeginRedelegate
func SimulateMsgBeginRedelegate(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgBeginRedelegate with %s", "ok")
		return msg, nil
	}
}

// SimulateMsgCompleteRedelegate
func SimulateMsgCompleteRedelegate(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgCompleteRedelegate with %s", "ok")
		return msg, nil
	}
}

// SimulationSetup
func SimulationSetup(mapp *mock.App, k Keeper) mock.RandSetup {
	return func(r *rand.Rand, privKeys []crypto.PrivKey) {
		ctx := mapp.NewContext(false, abci.Header{})
		InitGenesis(ctx, k, DefaultGenesisState())
	}
}

// TestStakeWithRandomMessages
func TestStakeWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterWire(mapp.Cdc)
	mapper := mapp.AccountMapper
	coinKeeper := bank.NewKeeper(mapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeKeeper := NewKeeper(mapp.Cdc, stakeKey, coinKeeper, DefaultCodespace)
	mapp.Router().AddRoute("stake", NewHandler(stakeKeeper))

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{stakeKey})
	if err != nil {
		panic(err)
	}

	mapp.SimpleRandomizedTestingFromSeed(
		t, 20, []mock.TestAndRunMsg{
			SimulateMsgCreateValidator(mapper, stakeKeeper),
			SimulateMsgEditValidator(stakeKeeper),
			SimulateMsgDelegate(stakeKeeper),
			SimulateMsgBeginUnbonding(stakeKeeper),
			SimulateMsgCompleteUnbonding(stakeKeeper),
			SimulateMsgBeginRedelegate(stakeKeeper),
			SimulateMsgCompleteRedelegate(stakeKeeper),
		}, []mock.RandSetup{
			SimulationSetup(mapp, stakeKeeper),
		}, []mock.Invariant{
			ModuleInvariants(coinKeeper, stakeKeeper),
		}, 10, 100, 500,
	)
}
