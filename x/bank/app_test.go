package bank

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"

	abci "github.com/tendermint/abci/types"
	crypto "github.com/tendermint/go-crypto"
)

// test bank module in a mock application
var (
	chainID = "" // TODO

	accName = "foobart"

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
	fee       = auth.StdFee{
		sdk.Coins{{"foocoin", 0}},
		100000,
	}

	sendMsg1 = MsgSend{
		Inputs:  []Input{NewInput(addr1, coins)},
		Outputs: []Output{NewOutput(addr2, coins)},
	}
)

func TestMsgSendWithAccounts(t *testing.T) {

	// initialize the mock application
	mapp := mock.NewApp()

	RegisterWire(mapp.Cdc)
	coinKeeper := NewKeeper(mapp.AccountMapper)
	mapp.Router().AddRoute("bank", NewHandler(coinKeeper))

	mapp.CompleteSetup(t, []*sdk.KVStoreKey{})

	// Add an account at genesis
	coins, err := sdk.ParseCoins("77foocoin")
	require.Nil(t, err)
	//acc := auth.NewAccountWithAddress(addr1)
	//acc.SetCoins(coins)
	//accs := []auth.Account{acc}

	baseAcc := &auth.BaseAccount{
		Address: addr1,
		Coins:   coins,
	}
	baseAccs := []auth.Account{baseAcc}

	// Construct genesis state
	mock.SetGenesis(mapp, baseAccs)

	// A checkTx context (true)
	ctxCheck := mapp.BaseApp.NewContext(true, abci.Header{})
	res1 := mapp.AccountMapper.GetAccount(ctxCheck, addr1)
	require.NotNil(t, res1)
	assert.Equal(t, baseAcc, res1.(*auth.BaseAccount))

	// Run a CheckDeliver
	mock.SignCheckDeliver(t, mapp, sendMsg1, []int64{0}, true, priv1)

	// Check balances
	mock.CheckBalance(t, mapp, addr1, sdk.Coins{{"foocoin", 67}})
	mock.CheckBalance(t, mapp, addr2, sdk.Coins{{"foocoin", 10}})

	// Delivering again should cause replay error
	mock.SignCheckDeliver(t, mapp, sendMsg1, []int64{0}, false, priv1)

	// bumping the txnonce number without resigning should be an auth error
	tx := mock.GenTx(sendMsg1, []int64{0}, priv1)
	tx.Signatures[0].Sequence = 1
	res := mapp.Deliver(tx)

	assert.Equal(t, sdk.ToABCICode(sdk.CodespaceRoot, sdk.CodeUnauthorized), res.Code, res.Log)

	// resigning the tx with the bumped sequence should work
	mock.SignCheckDeliver(t, mapp, sendMsg1, []int64{1}, true, priv1)
}
