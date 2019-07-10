package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// SimulateDeductFee
func SimulateDeductFee(ak auth.AccountKeeper, supplyKeeper types.SupplyKeeper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		accs []simulation.Account) (
		opMsg simulation.OperationMsg, fOps []simulation.FutureOperation, err error) {

		account := simulation.RandomAcc(r, accs)
		stored := ak.GetAccount(ctx, account.Address)
		initCoins := stored.GetCoins()
		opMsg = simulation.NewOperationMsgBasic(types.ModuleName, "deduct_fee", "", false, nil)

		feeCollector := ak.GetAccount(ctx, supplyKeeper.GetModuleAddress(types.FeeCollectorName))
		if feeCollector == nil {
			panic(fmt.Errorf("fee collector account hasn't been set"))
		}

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
		fees := sdk.NewCoins(sdk.NewCoin(randCoin.Denom, amt))
		spendableCoins := stored.SpendableCoins(ctx.BlockHeader().Time)
		if _, hasNeg := spendableCoins.SafeSub(fees); hasNeg {
			return opMsg, nil, nil
		}

		// get the new account balance
		_, hasNeg := initCoins.SafeSub(fees)
		if hasNeg {
			return opMsg, nil, nil
		}

		err = supplyKeeper.SendCoinsFromAccountToModule(ctx, stored.GetAddress(), types.FeeCollectorName, fees)
		if err != nil {
			panic(err)
		}

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
