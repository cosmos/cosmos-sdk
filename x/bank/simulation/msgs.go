package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/tendermint/tendermint/crypto"
)

// SimulateSingleInputMsgSend tests and runs a single msg send, with one input and one output, where both
// accounts already exist.
func SimulateSingleInputMsgSend(mapper auth.AccountMapper) simulation.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, event func(string)) (action string, fOps []simulation.FutureOperation, err error) {
		fromKey := simulation.RandomKey(r, keys)
		fromAddr := sdk.AccAddress(fromKey.PubKey().Address())
		toKey := simulation.RandomKey(r, keys)
		// Disallow sending money to yourself
		for {
			if !fromKey.Equals(toKey) {
				break
			}
			toKey = simulation.RandomKey(r, keys)
		}
		toAddr := sdk.AccAddress(toKey.PubKey().Address())
		initFromCoins := mapper.GetAccount(ctx, fromAddr).GetCoins()

		if len(initFromCoins) == 0 {
			return "skipping, no coins at all", nil, nil
		}

		denomIndex := r.Intn(len(initFromCoins))
		amt, goErr := randPositiveInt(r, initFromCoins[denomIndex].Amount)
		if goErr != nil {
			return "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, nil, nil
		}

		action = fmt.Sprintf("%s is sending %s %s to %s",
			fromAddr.String(),
			amt.String(),
			initFromCoins[denomIndex].Denom,
			toAddr.String(),
		)

		coins := sdk.Coins{{initFromCoins[denomIndex].Denom, amt}}
		var msg = bank.MsgSend{
			Inputs:  []bank.Input{bank.NewInput(fromAddr, coins)},
			Outputs: []bank.Output{bank.NewOutput(toAddr, coins)},
		}
		goErr = sendAndVerifyMsgSend(app, mapper, msg, ctx, []crypto.PrivKey{fromKey})
		if goErr != nil {
			return "", nil, goErr
		}
		event("bank/sendAndVerifyMsgSend/ok")

		return action, nil, nil
	}
}

// Sends and verifies the transition of a msg send. This fails if there are repeated inputs or outputs
func sendAndVerifyMsgSend(app *baseapp.BaseApp, mapper auth.AccountMapper, msg bank.MsgSend, ctx sdk.Context, privkeys []crypto.PrivKey) error {
	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	AccountNumbers := make([]int64, len(msg.Inputs))
	SequenceNumbers := make([]int64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := mapper.GetAccount(ctx, msg.Inputs[i].Address)
		AccountNumbers[i] = acc.GetAccountNumber()
		SequenceNumbers[i] = acc.GetSequence()
		initialInputAddrCoins[i] = acc.GetCoins()
	}
	for i := 0; i < len(msg.Outputs); i++ {
		acc := mapper.GetAccount(ctx, msg.Outputs[i].Address)
		initialOutputAddrCoins[i] = acc.GetCoins()
	}
	tx := mock.GenTx([]sdk.Msg{msg},
		AccountNumbers,
		SequenceNumbers,
		privkeys...)
	res := app.Deliver(tx)
	if !res.IsOK() {
		// TODO: Do this in a more 'canonical' way
		return fmt.Errorf("Deliver failed %v", res)
	}

	for i := 0; i < len(msg.Inputs); i++ {
		terminalInputCoins := mapper.GetAccount(ctx, msg.Inputs[i].Address).GetCoins()
		if !initialInputAddrCoins[i].Minus(msg.Inputs[i].Coins).IsEqual(terminalInputCoins) {
			return fmt.Errorf("input #%d had an incorrect amount of coins", i)
		}
	}
	for i := 0; i < len(msg.Outputs); i++ {
		terminalOutputCoins := mapper.GetAccount(ctx, msg.Outputs[i].Address).GetCoins()
		if !terminalOutputCoins.IsEqual(initialOutputAddrCoins[i].Plus(msg.Outputs[i].Coins)) {
			return fmt.Errorf("output #%d had an incorrect amount of coins", i)
		}
	}
	return nil
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
