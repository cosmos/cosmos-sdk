package bank

import (
	"testing"

	"github.com/stretchr/testify/require"

	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/mock"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

// test bank module in a mock application
var (
	priv1 = ed25519.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = ed25519.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	addr3 = sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	priv4 = ed25519.GenPrivKey()
	addr4 = sdk.AccAddress(priv4.PubKey().Address())

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

func TestBankWithRandomMessages(t *testing.T) {
	mapp := getMockApp(t)
	setup := func(r *rand.Rand, keys []crypto.PrivKey) {
		return
	}

	mapp.RandomizedTesting(
		t,
		[]mock.TestAndRunTx{TestAndRunSingleInputMsgSend},
		[]mock.RandSetup{setup},
		[]mock.Invariant{ModuleInvariants},
		100, 30, 30,
	)
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
