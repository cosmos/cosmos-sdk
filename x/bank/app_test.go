package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/mock"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
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
	coins     = sdk.Coins{{"foocoin", 10}}
	halfCoins = sdk.Coins{{"foocoin", 5}}
	manyCoins = sdk.Coins{{"foocoin", 1}, {"barcoin", 1}}

	freeFee = auth.StdFee{ // no fees for a buncha gas
		sdk.Coins{{"foocoin", 0}},
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
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	coinKeeper := NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("bank", NewHandler(coinKeeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{})
	return mapp
}

func TestMsgSendWithAccounts(t *testing.T) {
	mapp := getMockApp(t)

	// Add an account at genesis
	acc := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{{"foocoin", 67}},
	}
	accs := []auth.Account{acc}

	// Construct genesis state
	mock.SetGenesis(mapp, accs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.NotNil(t, res1)
	assert.Equal(t, acc, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg1, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 57}})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{{"foocoin", 10}})

	// Delivering again should cause replay error
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg1, []int64{0}, []int64{0}, false, priv1)

	// bumping the txnonce number without resigning should be an auth error
	mapp.BeginBlock(abci.RequestBeginBlock{})
	tx := mock.GenTx(sendMsg1, []int64{0}, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1
	res := mapp.Deliver(tx)

	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg1, []int64{0}, []int64{1}, true, priv1)
}

func TestMsgSendMultipleOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}

	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}
	accs := []auth.Account{acc1, acc2}

	mock.SetGenesis(mapp, accs)

	// Simulate a Block
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg2, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 32}})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{{"foocoin", 47}})
	mock.CheckBalance(t, mapp, addr3, sdk.Coins{{"foocoin", 5}})
}

func TestSengMsgMultipleInOut(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}
	acc2 := &auth.BaseAccount{
		Address: addr2,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}
	acc4 := &auth.BaseAccount{
		Address: addr4,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}
	accs := []auth.Account{acc1, acc2, acc4}

	mock.SetGenesis(mapp, accs)

	// CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg3, []int64{0, 2}, []int64{0, 0}, true, priv1, priv4)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 32}})
	mock.CheckBalance(t, mapp, addr4, sdk.Coins{{"foocoin", 32}})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{{"foocoin", 52}})
	mock.CheckBalance(t, mapp, addr3, sdk.Coins{{"foocoin", 10}})
}

func TestMsgSendDependent(t *testing.T) {
	mapp := getMockApp(t)

	acc1 := &auth.BaseAccount{
		Address: addr1,
		Coins:   sdk.Coins{{"foocoin", 42}},
	}
	accs := []auth.Account{acc1}

	mock.SetGenesis(mapp, accs)

	// CheckDeliver
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg1, []int64{0}, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 32}})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{{"foocoin", 10}})

	// Simulate a Block
	mock.SignCheckDeliver(t, mapp.BaseApp, sendMsg4, []int64{1}, []int64{0}, true, priv2)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 42}})
}
