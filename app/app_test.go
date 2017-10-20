package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/tmlibs/log"
)

// DefaultHandler for the tests (coin, roles, ibc)
func DefaultHandler(feeDenom string) sdk.Handler {
	// use the default stack
	r := roles.NewHandler()
	i := ibc.NewHandler()

	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
	).
		IBC(ibc.NewMiddleware()).
		Apps(
			roles.NewMiddleware(),
			fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
			stack.Checkpoint{OnDeliver: true},
		).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(r),
			stack.WrapHandler(i),
		)
}

//--------------------------------------------------------
// test environment is a list of input and output accounts

type appTest struct {
	t       *testing.T
	chainID string
	app     *BaseApp
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
func (at *appTest) baseTx(coins coin.Coins) sdk.Tx {
	in := []coin.TxInput{{Address: at.acctIn.Actor(), Coins: coins}}
	out := []coin.TxOutput{{Address: at.acctOut.Actor(), Coins: coins}}
	tx := coin.NewSendTx(in, out)
	return tx
}

func (at *appTest) signTx(tx sdk.Tx) sdk.Tx {
	stx := auth.NewMulti(tx)
	auth.Sign(stx, at.acctIn.Key)
	return stx.Wrap()
}

func (at *appTest) getTx(coins coin.Coins, sequence uint32) sdk.Tx {
	tx := at.baseTx(coins)
	tx = nonce.NewTx(sequence, []sdk.Actor{at.acctIn.Actor()}, tx)
	tx = base.NewChainTx(at.chainID, 0, tx)
	return at.signTx(tx)
}

func (at *appTest) feeTx(coins coin.Coins, toll coin.Coin, sequence uint32) sdk.Tx {
	tx := at.baseTx(coins)
	tx = fee.NewFee(tx, toll, at.acctIn.Actor())
	tx = nonce.NewTx(sequence, []sdk.Actor{at.acctIn.Actor()}, tx)
	tx = base.NewChainTx(at.chainID, 0, tx)
	return at.signTx(tx)
}

// set the account on the app through InitState
func (at *appTest) initAccount(acct *coin.AccountWithKey) {
	err := at.app.InitState("coin", "account", acct.MakeOption())
	require.Nil(at.t, err, "%+v", err)
}

// reset the in and out accs to be one account each with 7mycoin
func (at *appTest) reset() {
	at.acctIn = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})
	at.acctOut = coin.NewAccountWithKey(coin.Coins{{"mycoin", 7}})

	// Note: switch logger if you want to get more info
	logger := log.TestingLogger()
	// logger := log.NewTracingLogger(log.NewTMLogger(os.Stdout))

	store, err := NewStoreApp("app-test", "", 0, logger)
	require.Nil(at.t, err, "%+v", err)
	at.app = NewBaseApp(store, DefaultHandler("mycoin"), nil)

	err = at.app.InitState("base", "chain_id", at.chainID)
	require.Nil(at.t, err, "%+v", err)

	at.initAccount(at.acctIn)
	at.initAccount(at.acctOut)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

func getBalance(key sdk.Actor, store state.SimpleDB) (coin.Coins, error) {
	cspace := stack.PrefixedStore(coin.NameCoin, store)
	acct, err := coin.GetAccount(cspace, key)
	return acct.Coins, err
}

func getAddr(addr []byte, state state.SimpleDB) (coin.Coins, error) {
	actor := auth.SigPerm(addr)
	return getBalance(actor, state)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(t *testing.T, tx sdk.Tx, checkTx bool) (res abci.Result, diffIn, diffOut coin.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.acctIn.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.acctOut.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalIn, err := getBalance(at.acctIn.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	endBalOut, err := getBalance(at.acctOut.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	return res, endBalIn.Minus(initBalIn), endBalOut.Minus(initBalOut)
}

//--------------------------------------------------------

func TestInitState(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	logger := log.TestingLogger()
	store, err := NewStoreApp("app-test", "", 0, logger)
	require.Nil(err, "%+v", err)
	app := NewBaseApp(store, DefaultHandler("atom"), nil)

	//testing ChainID
	chainID := "testChain"
	err = app.InitState("base", "chain_id", chainID)
	require.Nil(err, "%+v", err)
	assert.EqualValues(app.GetChainID(), chainID)

	// make a nice account...
	bal := coin.Coins{{"atom", 77}, {"eth", 12}}
	acct := coin.NewAccountWithKey(bal)
	err = app.InitState("coin", "account", acct.MakeOption())
	require.Nil(err, "%+v", err)

	// make sure it is set correctly, with some balance
	coins, err := getBalance(acct.Actor(), app.Append())
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
	err = app.InitState("coin", "account", unsortAcc)
	require.Nil(err, "%+v", err)

	coins, err = getAddr(unsortAddr, app.Append())
	require.Nil(err)
	assert.True(coins.IsValid())
	assert.Equal(unsortCoins, coins)

	err = app.InitState("base", "dslfkgjdas", "")
	require.Error(err)

	err = app.InitState("", "dslfkgjdas", "")
	require.Error(err)

	err = app.InitState("dslfkgjdas", "szfdjzs", "")
	require.Error(err)
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

	key := stack.PrefixedKey(coin.NameCoin, at.acctIn.Address())
	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/key",
		Data: key,
	})
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}
