package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/txs"
	"github.com/tendermint/basecoin/types"
	crypto "github.com/tendermint/go-crypto"
	wire "github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/log"
)

//--------------------------------------------------------
// test environment is a list of input and output accounts

type appTest struct {
	t       *testing.T
	chainID string
	app     *Basecoin
	accIn   types.PrivAccount
	accOut  types.PrivAccount
}

func newAppTest(t *testing.T) *appTest {
	at := &appTest{
		t:       t,
		chainID: "test_chain_id",
	}
	at.reset()
	return at
}

// make a tx sending 5mycoin from each accIn to accOut
func (at *appTest) getTx(seq int, coins types.Coins) basecoin.Tx {
	addrIn := at.accIn.Account.PubKey.Address()
	addrOut := at.accOut.Account.PubKey.Address()

	in := []coin.TxInput{{Address: stack.SigPerm(addrIn), Coins: coins, Sequence: seq}}
	out := []coin.TxOutput{{Address: stack.SigPerm(addrOut), Coins: coins}}
	tx := coin.NewSendTx(in, out)
	tx = txs.NewChain(at.chainID, tx)
	stx := txs.NewMulti(tx)
	txs.Sign(stx, at.accIn.PrivKey)
	return stx.Wrap()
}

// set the account on the app through SetOption
func (at *appTest) acc2app(acc types.Account) {
	accBytes, err := json.Marshal(acc)
	require.Nil(at.t, err)
	res := at.app.SetOption("base/account", string(accBytes))
	require.EqualValues(at.t, res, "Success")
}

// reset the in and out accs to be one account each with 7mycoin
func (at *appTest) reset() {
	at.accIn = types.MakeAcc("input0")
	at.accOut = types.MakeAcc("output0")

	eyesCli := eyes.NewLocalClient("", 0)
	at.app = NewBasecoin(DefaultHandler(), eyesCli,
		log.TestingLogger().With("module", "app"))

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.acc2app(at.accIn.Account)
	at.acc2app(at.accOut.Account)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

func getBalance(pk crypto.PubKey, state types.KVStore) (types.Coins, error) {
	return getAddr(pk.Address(), state)
}

func getAddr(addr []byte, state types.KVStore) (types.Coins, error) {
	actor := stack.SigPerm(addr)
	acct, err := coin.NewAccountant("").GetAccount(state, actor)
	return acct.Coins, err
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(t *testing.T, tx basecoin.Tx, checkTx bool) (res abci.Result, diffIn, diffOut types.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.accIn.Account.PubKey, at.app.GetState())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.accOut.Account.PubKey, at.app.GetState())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalIn, err := getBalance(at.accIn.Account.PubKey, at.app.GetState())
	require.Nil(err, "%+v", err)
	endBalOut, err := getBalance(at.accOut.Account.PubKey, at.app.GetState())
	require.Nil(err, "%+v", err)
	return res, endBalIn.Minus(initBalIn), endBalOut.Minus(initBalOut)
}

//--------------------------------------------------------

func TestSetOption(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eyesCli := eyes.NewLocalClient("", 0)
	app := NewBasecoin(
		DefaultHandler(),
		eyesCli,
		log.TestingLogger().With("module", "app"),
	)

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	// make a nice account...
	accIn := types.MakeAcc("input0").Account
	accsInBytes, err := json.Marshal(accIn)
	assert.Nil(err)
	res = app.SetOption("base/account", string(accsInBytes))
	require.EqualValues(res, "Success")

	// make sure it is set correctly, with some balance
	coins, err := getBalance(accIn.PubKey, app.state)
	require.Nil(err)
	assert.Equal(accIn.Balance, coins)

	// let's parse an account with badly sorted coins...
	unsortAddr, err := hex.DecodeString("C471FB670E44D219EE6DF2FC284BE38793ACBCE1")
	require.Nil(err)
	unsortCoins := types.Coins{{"BTC", 789}, {"eth", 123}}
	unsortAcc := `{
  "pub_key": {
    "type": "ed25519",
    "data": "AD084F0572C116D618B36F2EB08240D1BAB4B51716CCE0E7734B89C8936DCE9A"
  },
  "coins": [
    {
      "denom": "eth",
      "amount": 123
    },
    {
      "denom": "BTC",
      "amount": 789
    }
  ]
}`
	res = app.SetOption("base/account", unsortAcc)
	require.EqualValues(res, "Success")

	coins, err = getAddr(unsortAddr, app.state)
	require.Nil(err)
	assert.True(coins.IsValid())
	assert.Equal(unsortCoins, coins)

	res = app.SetOption("base/dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas/szfdjzs", "")
	assert.NotEqual(res, "Success")

}

// Test CheckTx and DeliverTx with insufficient and sufficient balance
func TestTx(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	//Bad Balance
	at.accIn.Balance = types.Coins{{"mycoin", 2}}
	at.acc2app(at.accIn.Account)
	res, _, _ := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)
	res, diffIn, diffOut := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.True(diffIn.IsZero())
	assert.True(diffOut.IsZero())

	//Regular CheckTx
	at.reset()
	res, _, _ = at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	at.reset()
	amt := types.Coins{{"mycoin", 3}}
	res, diffIn, diffOut = at.exec(t, at.getTx(1, amt), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.Equal(amt.Negative(), diffIn)
	assert.Equal(amt, diffOut)
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _ := at.exec(t, at.getTx(1, types.Coins{{"mycoin", 5}}), false)
	assert.True(res.IsOK(), "Commit, DeliverTx: Expected OK return from DeliverTx, Error: %v", res)

	resQueryPreCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accIn.Account.PubKey.Address(),
	})

	res = at.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accIn.Account.PubKey.Address(),
	})
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}
