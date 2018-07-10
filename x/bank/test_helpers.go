package bank

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// ModuleInvariants runs all invariants of the bank module.
// Currently runs non-negative balance invariant and TotalCoinsInvariant
func ModuleInvariants(t *testing.T, app *mock.App, log string) {
	NonnegativeBalanceInvariant(t, app, log)
	TotalCoinsInvariant(t, app, log)
}

// NonnegativeBalanceInvariant checks that all accounts in the application have non-negative balances
func NonnegativeBalanceInvariant(t *testing.T, app *mock.App, log string) {
	ctx := app.NewContext(false, abci.Header{})
	accts := mock.GetAllAccounts(app.AccountMapper, ctx)
	for _, acc := range accts {
		coins := acc.GetCoins()
		assert.True(t, coins.IsNotNegative(),
			fmt.Sprintf("%s has a negative denomination of %s\n%s",
				acc.GetAddress().String(),
				coins.String(),
				log),
		)
	}
}

// TotalCoinsInvariant checks that the sum of the coins across all accounts
// is what is expected
func TotalCoinsInvariant(t *testing.T, app *mock.App, log string) {
	ctx := app.BaseApp.NewContext(false, abci.Header{})
	totalCoins := sdk.Coins{}

	chkAccount := func(acc auth.Account) bool {
		coins := acc.GetCoins()
		totalCoins = totalCoins.Plus(coins)
		return false
	}

	app.AccountMapper.IterateAccounts(ctx, chkAccount)
	require.Equal(t, app.TotalCoinsSupply, totalCoins, log)
}

// TestAndRunSingleInputMsgSend tests and runs a single msg send, with one input and one output, where both
// accounts already exist.
func TestAndRunSingleInputMsgSend(t *testing.T, r *rand.Rand, app *mock.App, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
	fromKey := keys[r.Intn(len(keys))]
	fromAddr := sdk.AccAddress(fromKey.PubKey().Address())
	toKey := keys[r.Intn(len(keys))]
	// Disallow sending money to yourself
	for {
		if !fromKey.Equals(toKey) {
			break
		}
		toKey = keys[r.Intn(len(keys))]
	}
	toAddr := sdk.AccAddress(toKey.PubKey().Address())
	initFromCoins := app.AccountMapper.GetAccount(ctx, fromAddr).GetCoins()

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
	var msg = MsgSend{
		Inputs:  []Input{NewInput(fromAddr, coins)},
		Outputs: []Output{NewOutput(toAddr, coins)},
	}
	sendAndVerifyMsgSend(t, app, msg, ctx, log, []crypto.PrivKey{fromKey})

	return action, nil
}

// Sends and verifies the transition of a msg send. This fails if there are repeated inputs or outputs
func sendAndVerifyMsgSend(t *testing.T, app *mock.App, msg MsgSend, ctx sdk.Context, log string, privkeys []crypto.PrivKey) {
	initialInputAddrCoins := make([]sdk.Coins, len(msg.Inputs))
	initialOutputAddrCoins := make([]sdk.Coins, len(msg.Outputs))
	AccountNumbers := make([]int64, len(msg.Inputs))
	SequenceNumbers := make([]int64, len(msg.Inputs))

	for i := 0; i < len(msg.Inputs); i++ {
		acc := app.AccountMapper.GetAccount(ctx, msg.Inputs[i].Address)
		AccountNumbers[i] = acc.GetAccountNumber()
		SequenceNumbers[i] = acc.GetSequence()
		initialInputAddrCoins[i] = acc.GetCoins()
	}
	for i := 0; i < len(msg.Outputs); i++ {
		acc := app.AccountMapper.GetAccount(ctx, msg.Outputs[i].Address)
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
		terminalInputCoins := app.AccountMapper.GetAccount(ctx, msg.Inputs[i].Address).GetCoins()
		require.Equal(t,
			initialInputAddrCoins[i].Minus(msg.Inputs[i].Coins),
			terminalInputCoins,
			fmt.Sprintf("Input #%d had an incorrect amount of coins\n%s", i, log),
		)
	}
	for i := 0; i < len(msg.Outputs); i++ {
		terminalOutputCoins := app.AccountMapper.GetAccount(ctx, msg.Outputs[i].Address).GetCoins()
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
