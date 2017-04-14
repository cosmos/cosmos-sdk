package app

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

//--------------------------------------------------------
// test environment is a list of input and output accounts

type appTest struct {
	t       *testing.T
	chainID string
	app     *Basecoin
	accsIn  []types.PrivAccount
	accsOut []types.PrivAccount
}

func newAppTest(t *testing.T) *appTest {
	at := &appTest{
		t:       t,
		chainID: "test_chain_id",
	}
	at.reset()
	return at
}

// make a tx sending 5mycoin from each accsIn to accsOut
func (ap *appTest) getTx(seq int) *types.SendTx {
	tx := types.GetTx(seq, ap.accsIn, ap.accsOut)
	types.SignTx(ap.chainID, tx, ap.accsIn)
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
	at.accsIn = types.MakeAccs("input0")
	at.accsOut = types.MakeAccs("output0")

	eyesCli := eyes.NewLocalClient("", 0)
	at.app = NewBasecoin(eyesCli)

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.acc2app(at.accsIn[0].Account)
	at.acc2app(at.accsOut[0].Account)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

// returns the final balance and expected balance for input and output accounts
func (at *appTest) exec(tx *types.SendTx, checkTx bool) (res abci.Result, inputGot, inputExp, outputGot, outputExpected types.Coins) {

	initBalFoo := at.app.GetState().GetAccount(at.accsIn[0].Account.PubKey.Address()).Balance
	initBalBar := at.app.GetState().GetAccount(at.accsOut[0].Account.PubKey.Address()).Balance

	txBytes := []byte(wire.BinaryBytes(struct{ types.Tx }{tx}))
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalFoo := at.app.GetState().GetAccount(at.accsIn[0].Account.PubKey.Address()).Balance
	endBalBar := at.app.GetState().GetAccount(at.accsOut[0].Account.PubKey.Address()).Balance
	decrBalFooExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee})
	return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(tx.Outputs[0].Coins)
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

	eyesCli := eyes.NewLocalClient("", 0)
	app := NewBasecoin(eyesCli)

	//testing ChainID
	chainID := "testChain"
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	accsIn := types.MakeAccs("input0")
	accsInBytes, err := json.Marshal(accsIn[0].Account)
	assert.Nil(err)
	res = app.SetOption("base/account", string(accsInBytes))
	assert.EqualValues(res, "Success")

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
	at.accsIn[0].Balance = types.Coins{{"mycoin", 2}}
	at.acc2app(at.accsIn[0].Account)
	res, _, _, _, _ := at.exec(at.getTx(1), true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, inGot, inExp, outGot, outExp := at.exec(at.getTx(1), false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.False(inGot.IsEqual(inExp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, inGot: %v, inExp: %v", inGot, inExp))
	assert.False(outGot.IsEqual(outExp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, outGot: %v, outExp: %v", outGot, outExp))

	//Regular CheckTx
	at.reset()
	res, _, _, _, _ = at.exec(at.getTx(1), true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	at.reset()
	res, inGot, inExp, outGot, outExp = at.exec(at.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(inGot.IsEqual(inExp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, inGot: %v, inExp: %v", inGot, inExp))
	assert.True(outGot.IsEqual(outExp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, outGot: %v, outExp: %v", outGot, outExp))
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _, _, _ := at.exec(at.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, DeliverTx: Expected OK return from DeliverTx, Error: %v", res))

	resQueryPreCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accsIn[0].Account.PubKey.Address(),
	})

	res = at.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accsIn[0].Account.PubKey.Address(),
	})
	fmt.Println(resQueryPreCommit)
	fmt.Println(resQueryPostCommit)
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}
