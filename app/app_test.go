package app

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
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
func (at *appTest) getTx(seq int) *types.SendTx {
	tx := types.GetTx(seq, at.accOut, at.accIn)
	types.SignTx(at.chainID, tx, at.accIn)
	return tx
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
	at.app = NewBasecoin(eyesCli)
	at.app.SetLogger(log.TestingLogger().With("module", "app"))

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.acc2app(at.accIn.Account)
	at.acc2app(at.accOut.Account)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(tx *types.SendTx, checkTx bool) (res abci.Result, inputGot, inputExp, outputGot, outputExpected types.Coins) {

	initBalIn := at.app.GetState().GetAccount(at.accIn.Account.PubKey.Address()).Balance
	initBalOut := at.app.GetState().GetAccount(at.accOut.Account.PubKey.Address()).Balance

	txBytes := []byte(wire.BinaryBytes(struct{ types.Tx }{tx}))
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalIn := at.app.GetState().GetAccount(at.accIn.Account.PubKey.Address()).Balance
	endBalOut := at.app.GetState().GetAccount(at.accOut.Account.PubKey.Address()).Balance
	decrBalInExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee})
	return res, endBalIn, initBalIn.Minus(decrBalInExp), endBalOut, initBalOut.Plus(tx.Outputs[0].Coins)
}

//--------------------------------------------------------

func TestSplitKey(t *testing.T) {
	assert := assert.New(t)
	prefix, suffix := splitKey("foo/bar")
	assert.EqualValues("foo", prefix)
	assert.EqualValues("bar", suffix)

	prefix, suffix = splitKey("foobar")
	assert.EqualValues("foobar", prefix)
	assert.EqualValues("", suffix)
}

func TestSetOption(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	eyesCli := eyes.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli)
	app.SetLogger(log.TestingLogger().With("module", "app"))

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	// make a nice account...
	accIn := types.MakeAcc("input0")
	accsInBytes, err := json.Marshal(accIn.Account)
	assert.Nil(err)
	res = app.SetOption("base/account", string(accsInBytes))
	require.EqualValues(res, "Success")
	// make sure it is set correctly, with some balance
	acct := state.GetAccount(app.GetState(), accIn.PubKey.Address())
	require.NotNil(acct)
	assert.Equal(accIn.Balance, acct.Balance)

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
	acct = state.GetAccount(app.GetState(), unsortAddr)
	require.NotNil(acct)
	assert.True(acct.Balance.IsValid())
	assert.Equal(unsortCoins, acct.Balance)

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
	res, _, _, _, _ := at.exec(at.getTx(1), true)
	assert.True(res.IsErr(), "ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res)
	res, inGot, inExp, outGot, outExp := at.exec(at.getTx(1), false)
	assert.True(res.IsErr(), "ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res)
	assert.False(inGot.IsEqual(inExp), "ExecTx/Bad DeliverTx: shouldn't be equal, inGot: %v, inExp: %v", inGot, inExp)
	assert.False(outGot.IsEqual(outExp), "ExecTx/Bad DeliverTx: shouldn't be equal, outGot: %v, outExp: %v", outGot, outExp)

	//Regular CheckTx
	at.reset()
	res, _, _, _, _ = at.exec(at.getTx(1), true)
	assert.True(res.IsOK(), "ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res)

	//Regular DeliverTx
	at.reset()
	res, inGot, inExp, outGot, outExp = at.exec(at.getTx(1), false)
	assert.True(res.IsOK(), "ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res)
	assert.True(inGot.IsEqual(inExp),
		"ExecTx/good DeliverTx: unexpected change in input coins, inGot: %v, inExp: %v", inGot, inExp)
	assert.True(outGot.IsEqual(outExp),
		"ExecTx/good DeliverTx: unexpected change in output coins, outGot: %v, outExp: %v", outGot, outExp)
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _, _, _ := at.exec(at.getTx(1), false)
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
