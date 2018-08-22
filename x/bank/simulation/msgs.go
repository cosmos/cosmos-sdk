package simulation

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/mock/simulation"
	"github.com/tendermint/tendermint/crypto"
)

// TestAndRunSingleInputMsgSend tests and runs a single msg send, with one input and one output, where both
// accounts already exist.
func TestAndRunSingleInputMsgSend(mapper auth.AccountMapper) simulation.TestAndRunTx {
	return func(t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, keys []crypto.PrivKey, log string, event func(string)) (action string, err sdk.Error) {
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
			return "skipping, no coins at all", nil
		}

		denomIndex := r.Intn(len(initFromCoins))
		amt, goErr := randPositiveInt(r, initFromCoins[denomIndex].Amount)
		if goErr != nil {
			return "skipping bank send due to account having no coins of denomination " + initFromCoins[denomIndex].Denom, nil
		}

		action = fmt.Sprintf("%s is sending %s %s to %s",
			fromAddr.String(),
			amt.String(),
			initFromCoins[denomIndex].Denom,
			toAddr.String(),
		)
		log = fmt.Sprintf("%s\n%s", log, action)

		coins := sdk.Coins{{initFromCoins[denomIndex].Denom, amt}}
		var msg = bank.MsgSend{
			Inputs:  []bank.Input{bank.NewInput(fromAddr, coins)},
			Outputs: []bank.Output{bank.NewOutput(toAddr, coins)},
		}
		sendAndVerifyMsgSend(t, app, mapper, msg, ctx, log, []crypto.PrivKey{fromKey})
		event("bank/sendAndVerifyMsgSend/ok")

		return action, nil
	}
}

// Sends and verifies the transition of a msg send. This fails if there are repeated inputs or outputs
func sendAndVerifyMsgSend(t *testing.T, app *baseapp.BaseApp, mapper auth.AccountMapper, msg bank.MsgSend, ctx sdk.Context, log string, privkeys []crypto.PrivKey) {
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
		fmt.Println(res)
		fmt.Println(log)
		t.FailNow()
	}

	for i := 0; i < len(msg.Inputs); i++ {
		terminalInputCoins := mapper.GetAccount(ctx, msg.Inputs[i].Address).GetCoins()
		require.Equal(t,
			initialInputAddrCoins[i].Minus(msg.Inputs[i].Coins),
			terminalInputCoins,
			fmt.Sprintf("Input #%d had an incorrect amount of coins\n%s", i, log),
		)
	}
	for i := 0; i < len(msg.Outputs); i++ {
		terminalOutputCoins := mapper.GetAccount(ctx, msg.Outputs[i].Address).GetCoins()
		require.Equal(t,
			initialOutputAddrCoins[i].Plus(msg.Outputs[i].Coins),
			terminalOutputCoins,
			fmt.Sprintf("Output #%d had an incorrect amount of coins\n%s", i, log),
		)
	}
}

func randPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GT(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}
	max = max.Sub(sdk.OneInt())
	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}
