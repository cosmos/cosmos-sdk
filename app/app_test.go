package app

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	eyes "github.com/tendermint/merkleeyes/client"
)

//TODO:

//Query -
//Commit - see if commit works before and after

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

//CheckTx - bad bytes, bad tx, good tx.
//DeliverTx - bad bytes, bad tx, good tx.
func TestTx(t *testing.T) {
	assert := assert.New(t)

	var accsFoo, accsBar []types.PrivAccount

	var app *Basecoin

	acc2app := func(acc types.Account) {
		accBytes, err := json.Marshal(acc)
		assert.Nil(err)
		res := app.SetOption("base/account", string(accBytes))
		assert.EqualValues(res, "Success")
	}

	reset := func() {
		accsFoo = makeAccs([]string{"foo"})
		accsBar = makeAccs([]string{"bar"})

		eyesCli := eyes.NewLocalClient("", 0)
		app = NewBasecoin(eyesCli)

		res := app.SetOption("base/chain_id", chainID)
		assert.EqualValues(res, "Success")

		acc2app(accsFoo[0].Account)
		acc2app(accsBar[0].Account)

		resabci := app.Commit()
		assert.True(resabci.IsOK(), resabci)
	}
	reset()

	accs2TxInputs := func(accs []types.PrivAccount, seq int) []types.TxInput {
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
	accs2TxOutputs := func(accs []types.PrivAccount) []types.TxOutput {
		var txs []types.TxOutput
		for _, acc := range accs {
			tx := types.TxOutput{
				acc.Account.PubKey.Address(),
				types.Coins{{"mycoin", 4}}}
			txs = append(txs, tx)
		}
		return txs
	}

	getTx := func(seq int) *types.SendTx {
		txs := &types.SendTx{
			Gas:     0,
			Fee:     types.Coin{"mycoin", 1},
			Inputs:  accs2TxInputs(accsFoo, seq),
			Outputs: accs2TxOutputs(accsBar),
		}
		signBytes := txs.SignBytes(chainID)
		for i, _ := range txs.Inputs {
			txs.Inputs[i].Signature = crypto.SignatureS{accsFoo[i].Sign(signBytes)}
		}

		return txs
	}
	txs := getTx(1)

	exec := func(checkTx bool) (res abci.Result, foo, fooExp, bar, barExp types.Coins) {

		initBalFoo := app.GetState().GetAccount(accsFoo[0].Account.PubKey.Address()).Balance
		initBalBar := app.GetState().GetAccount(accsBar[0].Account.PubKey.Address()).Balance

		txBytes := []byte(wire.BinaryBytes(struct {
			types.Tx `json:"unwrap"`
		}{txs}))

		if checkTx {
			res = app.CheckTx(txBytes)
		} else {
			res = app.DeliverTx(txBytes)
		}

		endBalFoo := app.GetState().GetAccount(accsFoo[0].Account.PubKey.Address()).Balance
		endBalBar := app.GetState().GetAccount(accsBar[0].Account.PubKey.Address()).Balance
		decrBalFooExp := txs.Outputs[0].Coins.Plus(types.Coins{txs.Fee})
		return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(txs.Outputs[0].Coins)
	}

	//Bad Balance
	accsFoo[0].Balance = types.Coins{{"mycoin", 2}}
	acc2app(accsFoo[0].Account)
	res, _, _, _, _ := exec(true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, foo, fooexp, bar, barexp := exec(false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.True(!foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(!bar.IsEqual(barexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, bar: %v, barExp: %v", bar, barexp))

	//Regular CheckTx
	reset()
	res, _, _, _, _ = exec(true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	reset()
	res, foo, fooexp, bar, barexp = exec(false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, bar: %v, barExp: %v", bar, barexp))

	///////////////////////
	//test Commit/Query
	//After Delivered TX foo should have no more coins to send,
	// but because the state hasn't yet been committed, checkTx should still
	// pass but after a commit it shouldn't

	reset()

	txs = getTx(1)
	res, _, _, _, _ = exec(false)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	txs = getTx(2)
	res, _, _, _, _ = exec(true)
	assert.True(res.IsOK(), fmt.Sprintf("Commit, CheckTx: Expected OK return from CheckTx, Error: %v", res))

	resQueryPreCommit := app.Query(abci.RequestQuery{
		Path: "/account",
		Data: accsFoo[0].Account.PubKey.Address(),
	})

	res = app.Commit()
	assert.True(res.IsOK(), res)

	resQueryPostCommit := app.Query(abci.RequestQuery{
		Path: "/account",
		Data: accsFoo[0].Account.PubKey.Address(),
	})
	fmt.Println(resQueryPreCommit)
	fmt.Println(resQueryPostCommit)

	assert.NotEqual(resQueryPreCommit, resQueryPostCommit, "Query should change before/after commit")
	txs = getTx(3)
	res, _, _, _, _ = exec(true)
	assert.True(res.IsErr(), fmt.Sprintf("Commit, CheckTx: Expected error return from CheckTx, returned: %v", res))

}
