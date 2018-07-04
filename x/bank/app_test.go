package bank

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"math/big"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

// test bank module in a mock application
var (
	priv1     = crypto.GenPrivKeyEd25519()
	addr1     = priv1.PubKey().Address()
	priv2     = crypto.GenPrivKeyEd25519()
	addr2     = priv2.PubKey().Address()
	addr3     = crypto.GenPrivKeyEd25519().PubKey().Address()
	priv4     = crypto.GenPrivKeyEd25519()
	addr4     = priv4.PubKey().Address()
	coins     = sdk.Coins{sdk.NewCoin("foocoin", 10)}
	halfCoins = sdk.Coins{sdk.NewCoin("foocoin", 5)}
	manyCoins = sdk.Coins{sdk.NewCoin("foocoin", 1), sdk.NewCoin("barcoin", 1)}

	freeFee = auth.StdFee{ // no fees for a buncha gas
		sdk.Coins{sdk.NewCoin("foocoin", 0)},
		100000,
	}

	sendMsg1 = MsgSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}

	sendMsg2 = MsgSend{
		Inputs: []Input{NewInput(addr1, coins)},
		Outputs: []Output{
			NewOutput(addr2, halfCoins),
			NewOutput(addr3, halfCoins),
		},
	}

	sendMsg3 = MsgSend{
		Inputs: []Input{
			NewInput(addr1, coins),
			NewInput(addr4, coins),
		},
		Outputs: []Output{
			NewOutput(addr2, coins),
			NewOutput(addr3, coins),
		},
	}

	sendMsg4 = MsgSend{
		Inputs: []Input{
			NewInput(addr2, coins),
		},
		Outputs: []Output{
			NewOutput(addr1, coins),
		},
	}

	sendMsg5 = MsgSend{
		Inputs: []Input{
			NewInput(addr1, manyCoins),
		},
		Outputs: []Output{
			NewOutput(addr2, manyCoins),
		},
	}
)

// initialize the mock application for this module
func getMockApp(t *testing.T) *mock.App {
	mapp, err := getBenchmarkMockApp()
	require.NoError(t, err)
	return mapp
}

// getBenchmarkMockApp initializes a mock application for this module, for purposes of benchmarking
// Any long term API support commitments do not apply to this function.
func getBenchmarkMockApp() (*mock.App, error) {
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	coinKeeper := NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("bank", NewHandler(coinKeeper))

	err := mapp.CompleteSetup([]*sdk.KVStoreKey{})
	return mapp, err
}

func TestBankWithRandomMessages(t *testing.T) {
	mapp := getMockApp(t)
	setup := func(r *rand.Rand, keys []crypto.PrivKey) {
		return
	}

	mapp.RandomizedTesting(
		t,
		[]mock.TestAndRunTx{randSingleSendTx},
		[]mock.RandSetup{setup},
		[]mock.AssertInvariants{bankTestInvariants(), mock.AuthInvariant},
		100, 30, 30,
	)
}

// Send a random "Send" Transaction from two already existing accounts
func randSingleSendTx(t *testing.T, r *rand.Rand, app *mock.App, ctx sdk.Context, keys []crypto.PrivKey, log string) (action string, err sdk.Error) {
	fromKey := keys[r.Intn(len(keys))]
	fromAddr := fromKey.PubKey().Address()
	toKey := keys[r.Intn(len(keys))]
	// Disallow sending money to yourself
	for {
		if !fromKey.Equals(toKey) {
			break
		}
		toKey = keys[r.Intn(len(keys))]
	}
	toAddr := toKey.PubKey().Address()
	initFromAcc := app.AccountMapper.GetAccount(ctx, fromAddr)
	initFromCoins := initFromAcc.GetCoins()

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
		Inputs:  []Input{NewInput(initFromAcc.GetAddress(), coins)},
		Outputs: []Output{NewOutput(toAddr, coins)},
	}
	sendAndVerifyMsgSend(t, app, msg, ctx, log, []crypto.PrivKey{fromKey})

	return action, nil
}

// Note this fails if there are repeated inputs or outputs
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
		// Do this the more 'canonical' way
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

func bankTestInvariants() mock.AssertInvariants {
	return func(t *testing.T, app *mock.App, log string) {
		// Check that noone has negative money
		ctx := app.NewContext(false, abci.Header{})
		checkNonnegativeBalances(t, app, ctx, log)
	}
}

func checkNonnegativeBalances(t *testing.T, app *mock.App, ctx sdk.Context, log string) {
	accts := mock.GetAllAccounts(app.AccountMapper, ctx)
	for _, acc := range accts {
		for _, coin := range acc.GetCoins() {
			assert.True(t, coin.IsNotNegative(), acc.GetAddress().String()+
				" has a negative denomination of "+coin.Denom+"\n"+log)
		}
	}
}

func TestMsgSendWithAccounts(t *testing.T) {
	mapp := getMockApp(t)

	// Add an account at genesis
	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 67)},
	}
	accs := []auth.Account{acc}

	// Construct genesis state
	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.NotNil(t, res1)
	require.Equal(t, acc, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1}, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 57)})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 10)})

	// Delivering again should cause replay error
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1, sendMsg2}, []int64{0}, []int64{0}, false, priv1)

	// bumping the txnonce number without resigning should be an auth error
	mapp.BeginBlock(abci.RequestBeginBlock{})
	tx := mock.GenTx([]sdk.Msg{sendMsg1}, []int64{0}, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1
	res := mapp.Deliver(tx)

	require.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1, sendMsg2}, []int64{0}, []int64{1}, true, priv1)
}

func TestMsgSendMultipleOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}

	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}
	accs := []auth.Account{acc1, acc2}

	mock.SetGenesis(mapp, accs)

	// Simulate a Block
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg2}, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 32)})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 47)})
	mock.CheckBalance(t, mapp, addr3, sdk.Coins{sdk.NewCoin("foocoin", 5)})
}

func TestSengMsgMultipleInOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}
	acc4 := &auth.BaseAccount{
		Address: addr4,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}
	accs := []auth.Account{acc1, acc2, acc4}

	mock.SetGenesis(mapp, accs)

	// CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg3}, []int64{0, 2}, []int64{0, 0}, true, priv1, priv4)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 32)})
	mock.CheckBalance(t, mapp, addr4, sdk.Coins{sdk.NewCoin("foocoin", 32)})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 52)})
	mock.CheckBalance(t, mapp, addr3, sdk.Coins{sdk.NewCoin("foocoin", 10)})
}

func TestMsgSendDependent(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{sdk.NewCoin("foocoin", 42)},
	}
	accs := []auth.Account{acc1}

	mock.SetGenesis(mapp, accs)

	// CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg1}, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 32)})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{sdk.NewCoin("foocoin", 10)})

	// Simulate a Block
	mock.SignCheckDeliver(t, mapp.BaseApp, []sdk.Msg{sendMsg4}, []int64{1}, []int64{0}, true, priv2)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{sdk.NewCoin("foocoin", 42)})
}
