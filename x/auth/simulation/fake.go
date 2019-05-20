package simulation

import (
	"errors"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateDeductFee
func SimulateDeductFee(m auth.AccountKeeper, f auth.FeeCollectionKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		account := simulation.RandomAcc(r, accs)
		stored := m.GetAccount(ctx, account.Address)
		initCoins := stored.GetCoins()
		opMsg = simulation.NewOperationMsgBasic(auth.ModuleName, "deduct_fee", "", false, nil)

		if len(initCoins) == 0 {
			return opMsg, nil, nil
		}

		denomIndex := r.Intn(len(initCoins))
		randCoin := initCoins[denomIndex]

		amt, err := randPositiveInt(r, randCoin.Amount)
		if err != nil {
			return opMsg, nil, nil
		}

		// Create a random fee and verify the fees are within the account's spendable
		// balance.
		fees := sdk.Coins{sdk.NewCoin(randCoin.Denom, amt)}
		spendableCoins := stored.SpendableCoins(ctx.BlockHeader().Time)
		if _, hasNeg := spendableCoins.SafeSub(fees); hasNeg {
			return opMsg, nil, nil
		}

		// get the new account balance
		newCoins, hasNeg := initCoins.SafeSub(fees)
		if hasNeg {
			return opMsg, nil, nil
		}

		if err := stored.SetCoins(newCoins); err != nil {
			panic(err)
		}

		m.SetAccount(ctx, stored)
		f.AddCollectedFees(ctx, fees)

		opMsg.OK = true
		return opMsg, nil, nil
	}
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
