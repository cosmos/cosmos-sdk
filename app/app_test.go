package app

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

/////////////////////
// Testing functions

func makeAccs(secrets []string) (accs []types.PrivAccount) {
	for _, secret := range secrets {
		privAcc := types.PrivAccountFromSecret(secret)
		privAcc.Account.Balance = types.Coins{{"mycoin", 7}}
		accs = append(accs, privAcc)
	}
	return
}

const chainID = "testChain"

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
	res := app.SetOption("base/chain_id", chainID)
	assert.EqualValues(app.GetState().GetChainID(), chainID)
	assert.EqualValues(res, "Success")

	accsFoo := makeAccs([]string{"foo"})
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

//////////////////////// TxTest

type testValues struct {
	t       *testing.T
	app     *Basecoin
	accsFoo []types.PrivAccount
	accsBar []types.PrivAccount
}

func (tv *testValues) acc2app(acc types.Account) {
	accBytes, err := json.Marshal(acc)
	require.Nil(tv.t, err)
	res := tv.app.SetOption("base/account", string(accBytes))
	require.EqualValues(tv.t, res, "Success")
}

func (tv *testValues) appInit() {
	tv.accsFoo = makeAccs([]string{"foo"})
	tv.accsBar = makeAccs([]string{"bar"})

	eyesCli := eyes.NewLocalClient("", 0)
	tv.app = NewBasecoin(eyesCli)

	res := tv.app.SetOption("base/chain_id", chainID)
	require.EqualValues(tv.t, res, "Success")

	tv.acc2app(tv.accsFoo[0].Account)
	tv.acc2app(tv.accsBar[0].Account)

	resabci := tv.app.Commit()
	require.True(tv.t, resabci.IsOK(), resabci)
}

func accs2TxInputs(accs []types.PrivAccount, seq int) []types.TxInput {
	var txs []types.TxInput
	for _, acc := range accs {
		tx := types.NewTxInput(
			acc.Account.PubKey,
			types.Coins{{"mycoin", 5}},
			seq)
		txs = append(txs, tx)
	}
	return txs
}

//turn a list of accounts into basic list of transaction outputs
func accs2TxOutputs(accs []types.PrivAccount) []types.TxOutput {
	var txs []types.TxOutput
	for _, acc := range accs {
		tx := types.TxOutput{
			acc.Account.PubKey.Address(),
			types.Coins{{"mycoin", 4}}}
		txs = append(txs, tx)
	}
	return txs
}

func (tv testValues) getTx(seq int) *types.SendTx {
	txs := &types.SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Inputs:  accs2TxInputs(tv.accsFoo, seq),
		Outputs: accs2TxOutputs(tv.accsBar),
	}
	signBytes := txs.SignBytes(chainID)
	for i, _ := range txs.Inputs {
		txs.Inputs[i].Signature = crypto.SignatureS{tv.accsFoo[i].Sign(signBytes)}
	}

	return txs
}

func (tv testValues) exec(tx *types.SendTx, checkTx bool) (res abci.Result, foo, fooExp, bar, barExp types.Coins) {

	initBalFoo := tv.app.GetState().GetAccount(tv.accsFoo[0].Account.PubKey.Address()).Balance
	initBalBar := tv.app.GetState().GetAccount(tv.accsBar[0].Account.PubKey.Address()).Balance

	txBytes := []byte(wire.BinaryBytes(struct {
		types.Tx `json:"unwrap"`
	}{tx}))

	if checkTx {
		res = tv.app.CheckTx(txBytes)
	} else {
		res = tv.app.DeliverTx(txBytes)
	}

	endBalFoo := tv.app.GetState().GetAccount(tv.accsFoo[0].Account.PubKey.Address()).Balance
	endBalBar := tv.app.GetState().GetAccount(tv.accsBar[0].Account.PubKey.Address()).Balance
	decrBalFooExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee})
	return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(tx.Outputs[0].Coins)
}

//CheckTx - bad bytes, bad tx, good tx.
//DeliverTx - bad bytes, bad tx, good tx.
func TestTx(t *testing.T) {
	assert := assert.New(t)

	tv := testValues{t: t}
	tv.appInit()

	//Bad Balance
	tv.accsFoo[0].Balance = types.Coins{{"mycoin", 2}}
	tv.acc2app(tv.accsFoo[0].Account)
	res, _, _, _, _ := tv.exec(tv.getTx(1), true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, foo, fooexp, bar, barexp := tv.exec(tv.getTx(1), false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.True(!foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(!bar.IsEqual(barexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, bar: %v, barExp: %v", bar, barexp))

	//Regular CheckTx
	tv.appInit()
	res, _, _, _, _ = tv.exec(tv.getTx(1), true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	tv.appInit()
	res, foo, fooexp, bar, barexp = tv.exec(tv.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, bar: %v, barExp: %v", bar, barexp))
}

func TestQuery(t *testing.T) {
	assert := assert.New(t)
	tv := testValues{t: t}
	tv.appInit()

	res, _, _, _, _ := tv.exec(tv.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	resQueryPreCommit := tv.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: tv.accsFoo[0].Account.PubKey.Address(),
	})

	res = tv.app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := tv.app.Query(abci.RequestQuery{
		Path: "/account",
		Data: tv.accsFoo[0].Account.PubKey.Address(),
	})
	fmt.Println(resQueryPreCommit)
	fmt.Println(resQueryPostCommit)
	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
}

func TestCommit(t *testing.T) {
	assert := assert.New(t)
	tv := testValues{t: t}
	tv.appInit()

	//After Delivered TX foo should have no more coins to send,
	// but because the state hasn't yet been committed, checkTx should still
	// pass but after a commit it shouldn't
	res, _, _, _, _ := tv.exec(tv.getTx(1), false)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	res, _, _, _, _ = tv.exec(tv.getTx(2), true)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	res = tv.app.Commit()
	assert.True(res.IsOK(), res)

	res, _, _, _, _ = tv.exec(tv.getTx(3), true)
	assert.True(res.IsErr(), fmt.Sprintf("Commit, CheckTx: Expected error return from CheckTx, returned: %v", res))
}
