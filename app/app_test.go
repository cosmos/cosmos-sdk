package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/modules/auth"
	"github.com/tendermint/basecoin/modules/base"
	"github.com/tendermint/basecoin/modules/coin"
	"github.com/tendermint/basecoin/modules/fee"
	"github.com/tendermint/basecoin/modules/nonce"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/state/merkle"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"
)

//--------------------------------------------------------
// test environment is a list of input and output accounts

type appTest struct {
	t       *testing.T
	chainID string
	app     *Basecoin
	acctIn  *coin.AccountWithKey
	acctOut *coin.AccountWithKey
}

func newAppTest(t *testing.T) *appTest {
	at := &appTest{
		t:       t,
		chainID: "test_chain_id",
	}
	at.reset()
	return at
}

// baseTx is the
func (at *appTest) baseTx(coins coin.Coins) basecoin.Tx {
	in := []coin.TxInput{{Address: at.acctIn.Actor(), Coins: coins}}
	out := []coin.TxOutput{{Address: at.acctOut.Actor(), Coins: coins}}
	tx := coin.NewSendTx(in, out)
	return tx
}

func (at *appTest) signTx(tx basecoin.Tx) basecoin.Tx {
	stx := auth.NewMulti(tx)
	auth.Sign(stx, at.acctIn.Key)
	return stx.Wrap()
}

func (at *appTest) getTx(coins coin.Coins, sequence uint32) basecoin.Tx {
	tx := at.baseTx(coins)
	tx = nonce.NewTx(sequence, []basecoin.Actor{at.acctIn.Actor()}, tx)
	tx = base.NewChainTx(at.chainID, 0, tx)
	return at.signTx(tx)
}

func (at *appTest) feeTx(coins coin.Coins, toll coin.Coin, sequence uint32) basecoin.Tx {
	tx := at.baseTx(coins)
	tx = fee.NewFee(tx, toll, at.acctIn.Actor())
	tx = nonce.NewTx(sequence, []basecoin.Actor{at.acctIn.Actor()}, tx)
	tx = base.NewChainTx(at.chainID, 0, tx)
	return at.signTx(tx)
}

// set the account on the app through SetOption
func (at *appTest) initAccount(acct *coin.AccountWithKey) {
	res := at.app.SetOption("coin/account", acct.MakeOption())
	require.EqualValues(at.t, res, "Success")
}

// reset the in and out accs to be one account each with 7mycoin
func (at *appTest) reset() {
	at.acctIn = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})
	at.acctOut = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})

	// Note: switch logger if you want to get more info
	logger := log.TestingLogger()
	// logger := log.NewTracingLogger(log.NewTMLogger(os.Stdout))
	store := merkle.NewStore("", 0, logger.With("module", "store"))
	at.app = NewBasecoin(
		DefaultHandler("mycoin"),
		store,
		logger.With("module", "app"),
	)

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.initAccount(at.acctIn)
	at.initAccount(at.acctOut)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

func getBalance(key basecoin.Actor, store state.KVStore) (coin.Coins, error) {
	cspace := stack.PrefixedStore(coin.NameCoin, store)
	acct, err := coin.GetAccount(cspace, key)
	return acct.Coins, err
}

func getAddr(addr []byte, state state.KVStore) (coin.Coins, error) {
	actor := auth.SigPerm(addr)
	return getBalance(actor, state)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(t *testing.T, tx basecoin.Tx, checkTx bool) (res abci.Result, diffIn, diffOut coin.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.acctIn.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.acctOut.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalIn, err := getBalance(at.acctIn.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	endBalOut, err := getBalance(at.acctOut.Actor(), at.app.GetState())
	require.Nil(err, "%+v", err)
	return res, endBalIn.Minus(initBalIn), endBalOut.Minus(initBalOut)
}

//--------------------------------------------------------

func TestSetOption(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	logger := log.TestingLogger()
	store := merkle.NewStore("", 0, logger.With("module", "store"))
	app := NewBasecoin(
		DefaultHandler("atom"),
		store,
		logger.With("module", "app"),
	)

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	// make a nice account...
	bal := coin.Coins{{"atom", 77}, {"eth", 12}}
	acct := coin.NewAccountWithKey(bal)
	res = app.SetOption("coin/account", acct.MakeOption())
	require.EqualValues(res, "Success")

	// make sure it is set correctly, with some balance
	coins, err := getBalance(acct.Actor(), app.GetState())
	require.Nil(err)
	assert.Equal(bal, coins)

	// let's parse an account with badly sorted coins...
	unsortAddr, err := hex.DecodeString("C471FB670E44D219EE6DF2FC284BE38793ACBCE1")
	require.Nil(err)
	unsortCoins := coin.Coins{{"BTC", 789}, {"eth", 123}}
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
	res = app.SetOption("coin/account", unsortAcc)
	require.EqualValues(res, "Success")

	coins, err = getAddr(unsortAddr, app.GetState())
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
	at.acctIn.Coins = coin.Coins{{"mycoin", 2}}
	at.initAccount(at.acctIn)
	at.app.Commit()

	res, _, _ := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1), true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)
	res, diffIn, diffOut := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1), false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.True(diffIn.IsZero())
	assert.True(diffOut.IsZero())

	//Regular CheckTx
	at.reset()
	res, _, _ = at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1), true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	at.reset()
	amt := coin.Coins{{"mycoin", 3}}
	res, diffIn, diffOut = at.exec(t, at.getTx(amt, 1), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.Equal(amt.Negative(), diffIn)
	assert.Equal(amt, diffOut)

	//DeliverTx with fee.... 4 get to recipient, 1 extra taxed
	at.reset()
	amt = coin.Coins{{"mycoin", 4}}
	toll := coin.Coin{"mycoin", 1}
	res, diffIn, diffOut = at.exec(t, at.feeTx(amt, toll, 1), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	payment := amt.Plus(coin.Coins{toll}).Negative()
	assert.Equal(payment, diffIn)
	assert.Equal(amt, diffOut)

}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _ := at.exec(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1), false)
	assert.True(res.IsOK(), "Commit, DeliverTx: Expected OK return from DeliverTx, Error: %v", res)

	resQueryPreCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.acctIn.Address(),
	})

	res = at.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.acctIn.Address(),
	})
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}

func TestSplitKey(t *testing.T) {
	assert := assert.New(t)
	prefix, suffix := splitKey("foo/bar")
	assert.EqualValues("foo", prefix)
	assert.EqualValues("bar", suffix)

	prefix, suffix = splitKey("foobar")
	assert.EqualValues("base", prefix)
	assert.EqualValues("foobar", suffix)

	prefix, suffix = splitKey("some/complex/issue")
	assert.EqualValues("some", prefix)
	assert.EqualValues("complex/issue", suffix)

}
