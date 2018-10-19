package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
)

// SimulateDeductFee
func SimulateDeductFee(m auth.AccountMapper, f auth.FeeCollectionKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account, event func(string)) (
		action string, fOp []simulation.FutureOperation, err error) {

		account := simulation.RandomAcc(r, accs)
		stored := m.GetAccount(ctx, account.Address)
		initCoins := stored.GetCoins()

		if len(initCoins) == 0 {
			event(fmt.Sprintf("auth/SimulateDeductFee/false"))
			return action, nil, nil
		}

		denomIndex := r.Intn(len(initCoins))
		amt, err := randPositiveInt(r, initCoins[denomIndex].Amount)
		if err != nil {
			event(fmt.Sprintf("auth/SimulateDeductFee/false"))
			return action, nil, nil
		}

		coins := sdk.Coins{sdk.NewCoin(initCoins[denomIndex].Denom, amt)}
		stored.SetCoins(initCoins.Minus(coins))
		m.SetAccount(ctx, stored)
		f.AddCollectedFees(ctx, coins)

		event(fmt.Sprintf("auth/SimulateDeductFee/true"))

		action = "TestDeductFee"
		return action, nil, nil
	}
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
