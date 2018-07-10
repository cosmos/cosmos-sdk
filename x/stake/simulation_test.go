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
		TotalSupplyInvariant(ck, k)(t, app, log)
		PositivePowerInvariant(k)(t, app, log)
	}
}

// TotalSupplyInvariant checks that the total supply reflects all held loose tokens, bonded tokens, and unbonding delegations
func TotalSupplyInvariant(ck bank.Keeper, k Keeper) mock.Invariant {
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

// SimulateMsgDelegate
func SimulateMsgDelegate(k Keeper) mock.TestAndRunMsg {
	return func(t *testing.T, r *rand.Rand, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
		msg := fmt.Sprintf("TestMsgDelegate with %s", "ok")
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

// Test random messages
func TestStakeWithRandomMessages(t *testing.T) {
	mapp := mock.NewApp()

	bank.RegisterWire(mapp.Cdc)
	coinKeeper := bank.NewKeeper(mapp.AccountMapper)
	stakeKey := sdk.NewKVStoreKey("stake")
	stakeKeeper := NewKeeper(mapp.Cdc, stakeKey, coinKeeper, DefaultCodespace)
	mapp.Router().AddRoute("stake", NewHandler(stakeKeeper))

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{stakeKey})
	if err != nil {
		panic(err)
	}

	mapp.SimpleRandomizedTestingFromSeed(
		t, 20, []mock.TestAndRunMsg{
			SimulateMsgDelegate(stakeKeeper),
		}, []mock.RandSetup{
			SimulationSetup(mapp, stakeKeeper),
		}, []mock.Invariant{
			ModuleInvariants(coinKeeper, stakeKeeper),
		}, 10, 100, 500,
	)
}
