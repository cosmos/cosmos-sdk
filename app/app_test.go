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

	accsFoo := types.MakeAccs("foo")
	accsFooBytes, err := json.Marshal(accsFoo[0].Account)
	assert.Nil(err)
	res = app.SetOption("base/account", string(accsFooBytes))
	assert.EqualValues(res, "Success")

	res = app.SetOption("base/dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas", "")
	assert.NotEqual(res, "Success")

	res = app.SetOption("dslfkgjdas/szfdjzs", "")
	assert.NotEqual(res, "Success")
}

//CheckTx - bad bytes, bad tx, good tx.
//DeliverTx - bad bytes, bad tx, good tx.
func TestTx(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	//Bad Balance
	at.accsFoo[0].Balance = types.Coins{{"mycoin", 2}}
	at.acc2app(at.accsFoo[0].Account)
	res, _, _, _, _ := at.exec(at.getTx(1), true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, foo, fooexp, bar, barexp := at.exec(at.getTx(1), false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.True(!foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(!bar.IsEqual(barexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, bar: %v, barExp: %v", bar, barexp))

	//Regular CheckTx
	at.reset()
	res, _, _, _, _ = at.exec(at.getTx(1), true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	at.reset()
	res, foo, fooexp, bar, barexp = at.exec(at.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, bar: %v, barExp: %v", bar, barexp))
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	at := newAppTest(t)

	res, _, _, _, _ := at.exec(at.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	resQueryPreCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accsFoo[0].Account.PubKey.Address(),
	})

	res = at.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := at.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: at.accsFoo[0].Account.PubKey.Address(),
	})
	fmt.Println(resQueryPreCommit)
	fmt.Println(resQueryPostCommit)
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}

/////////////////////////////////////////////////////////////////

type appTest struct {
	t       *testing.T
	chainID string
	app     *Basecoin
	accsFoo []types.PrivAccount
	accsBar []types.PrivAccount
}

func newAppTest(t *testing.T) *appTest {
	at := &appTest{
		t:       t,
		chainID: "test_chain_id",
	}
	at.reset()
	return at
}

func (ap *appTest) getTx(seq int) *types.SendTx {
	tx := types.GetTx(seq, ap.accsFoo, ap.accsBar)
	types.SignTx(ap.chainID, tx, ap.accsFoo)
	return tx
}

func (at *appTest) acc2app(acc types.Account) {
	accBytes, err := json.Marshal(acc)
	require.Nil(at.t, err)
	res := at.app.SetOption("base/account", string(accBytes))
	require.EqualValues(at.t, res, "Success")
}

func (at *appTest) reset() {
	at.accsFoo = types.MakeAccs("foo")
	at.accsBar = types.MakeAccs("bar")

	eyesCli := eyes.NewLocalClient("", 0)
	at.app = NewBasecoin(eyesCli)

	res := at.app.SetOption("base/chain_id", at.chainID)
	require.EqualValues(at.t, res, "Success")

	at.acc2app(at.accsFoo[0].Account)
	at.acc2app(at.accsBar[0].Account)

	resabci := at.app.Commit()
	require.True(at.t, resabci.IsOK(), resabci)
}

func (at *appTest) exec(tx *types.SendTx, checkTx bool) (res abci.Result, foo, fooExp, bar, barExp types.Coins) {

	initBalFoo := at.app.GetState().GetAccount(at.accsFoo[0].Account.PubKey.Address()).Balance
	initBalBar := at.app.GetState().GetAccount(at.accsBar[0].Account.PubKey.Address()).Balance

	txBytes := []byte(wire.BinaryBytes(struct{ types.Tx }{tx}))
	if checkTx {
		res = at.app.CheckTx(txBytes)
	} else {
		res = at.app.DeliverTx(txBytes)
	}

	endBalFoo := at.app.GetState().GetAccount(at.accsFoo[0].Account.PubKey.Address()).Balance
	endBalBar := at.app.GetState().GetAccount(at.accsBar[0].Account.PubKey.Address()).Balance
	decrBalFooExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee})
	return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(tx.Outputs[0].Coins)
}
