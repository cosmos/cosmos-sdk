package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	wire "github.com/tendermint/go-wire"
	//"github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/modules/auth"
	"github.com/cosmos/cosmos-sdk/modules/base"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/modules/fee"
	"github.com/cosmos/cosmos-sdk/modules/nonce"
	"github.com/cosmos/cosmos-sdk/modules/roles"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/cosmos-sdk/state"
)

// DefaultHandler for the tests (coin, roles, ibc)
func DefaultHandler(feeDenom string) sdk.Handler {
	// use the default stack
	r := roles.NewHandler()

	return stack.New(
		base.Logger{},
		stack.Recovery{},
		auth.Signatures{},
		base.Chain{},
		stack.Checkpoint{OnCheck: true},
		nonce.ReplayCheck{},
		roles.NewMiddleware(),
		fee.NewSimpleFeeMiddleware(coin.Coin{feeDenom, 0}, fee.Bank),
		stack.Checkpoint{OnDeliver: true},
	).
		Dispatch(
			coin.NewHandler(),
			stack.WrapHandler(r),
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
	_ = resabci
	//require.True(at.t, resabci.IsOK(), resabci)
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
func (at *appTest) execDeliver(t *testing.T, tx sdk.Tx) (res abci.ResponseDeliverTx, diffIn, diffOut coin.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.acctIn.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.acctOut.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	res = at.app.DeliverTx(txBytes)

	// check the tags
	if res.IsOK() {
		tags := res.Tags
		require.NotEmpty(tags)
		require.Equal("height", string(tags[0].GetKey()))
		//require.True(tags[0].GetValue() > 0)
		require.Equal("coin.sender", string(tags[1].GetKey()))
		//sender := at.acctIn.Actor().Address.String()
		//bz := common.HexBytes(tags[1].GetValue())
		//require.Equal(sender, bz.String())
		require.Equal("coin.receiver", string(tags[2].GetKey()))
		//rcpt := at.acctOut.Actor().Address.String()
		//bz = common.HexBytes(tags[2].GetValue())
		//require.Equal(rcpt, bz.String())
	}

	endBalIn, err := getBalance(at.acctIn.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	endBalOut, err := getBalance(at.acctOut.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	return res, endBalIn.Minus(initBalIn), endBalOut.Minus(initBalOut)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) execCheck(t *testing.T, tx sdk.Tx) (res abci.ResponseCheckTx, diffIn, diffOut coin.Coins) {
	require := require.New(t)

	initBalIn, err := getBalance(at.acctIn.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)
	initBalOut, err := getBalance(at.acctOut.Actor(), at.app.Append())
	require.Nil(err, "%+v", err)

	txBytes := wire.BinaryBytes(tx)
	res = at.app.CheckTx(txBytes)

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

	cres, _, _ := at.execCheck(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1))
	assert.True(cres.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", cres)
	dres, diffIn, diffOut := at.execDeliver(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1))
	assert.True(dres.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", dres)
	assert.True(diffIn.IsZero())
	assert.True(diffOut.IsZero())

	//Regular CheckTx
	at.reset()
	cres, _, _ = at.execCheck(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1))
	assert.True(cres.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", cres)

	//Regular DeliverTx
	at.reset()
	amt := coin.Coins{{"mycoin", 3}}
	dres, diffIn, diffOut = at.execDeliver(t, at.getTx(amt, 1))
	assert.True(dres.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", dres)
	assert.Equal(amt.Negative(), diffIn)
	assert.Equal(amt, diffOut)

	//DeliverTx with fee.... 4 get to recipient, 1 extra taxed
	at.reset()
	amt = coin.Coins{{"mycoin", 4}}
	toll := coin.Coin{"mycoin", 1}
	dres, diffIn, diffOut = at.execDeliver(t, at.feeTx(amt, toll, 1))
	assert.True(dres.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", dres)
	payment := amt.Plus(coin.Coins{toll}).Negative()
	assert.Equal(payment, diffIn)
	assert.Equal(amt, diffOut)

}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	dres, _, _ := at.execDeliver(t, at.getTx(coin.Coins{{"mycoin", 5}}, 1))
	assert.True(dres.IsOK(), "Commit, DeliverTx: Expected OK return from DeliverTx, Error: %v", dres)

	resQueryPreCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.acctIn.Address(),
	})

	cres := at.app.Commit()
	_ = cres
	//assert.True(cres.IsOK(), cres)

	key := stack.PrefixedKey(coin.NameCoin, at.acctIn.Address())
	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/key",
		Data: key,
	})
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}
