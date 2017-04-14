package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
)

//--------------------------------------------------------
// test environment is a bunch of lists of accountns

type execTest struct {
	chainID    string
	store      types.KVStore
	state      *State
	accsFoo    []types.PrivAccount
	accsBar    []types.PrivAccount
	accsFooBar []types.PrivAccount
	accsDup    []types.PrivAccount
}

func newExecTest() *execTest {
	et := &execTest{
		chainID: "test_chain_id",
	}
	et.reset()
	return et
}

func (et *execTest) signTx(tx *types.SendTx, accsIn []types.PrivAccount) {
	types.SignTx(et.chainID, tx, accsIn)
}

// make tx from accsIn to et.accsBar
func (et *execTest) getTx(seq int, accsIn []types.PrivAccount) *types.SendTx {
	return types.GetTx(seq, accsIn, et.accsBar)
}

// returns the final balance and expected balance for input and output accounts
func (et *execTest) exec(tx *types.SendTx, checkTx bool) (res abci.Result, inGot, inExp, outGot, outExp types.Coins) {
	initBalFoo := et.state.GetAccount(et.accsFoo[0].Account.PubKey.Address()).Balance
	initBalBar := et.state.GetAccount(et.accsBar[0].Account.PubKey.Address()).Balance

	res = ExecTx(et.state, nil, tx, checkTx, nil)

	endBalFoo := et.state.GetAccount(et.accsFoo[0].Account.PubKey.Address()).Balance
	endBalBar := et.state.GetAccount(et.accsBar[0].Account.PubKey.Address()).Balance
	decrBalFooExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee})
	return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(tx.Outputs[0].Coins)
}

func (et *execTest) acc2State(accs []types.PrivAccount) {
	for _, acc := range accs {
		et.state.SetAccount(acc.Account.PubKey.Address(), &acc.Account)
	}
}

//reset everything. state is empty
func (et *execTest) reset() {
	et.accsFoo = types.MakeAccs("foo")
	et.accsBar = types.MakeAccs("bar")
	et.accsFooBar = types.MakeAccs("foo", "bar")
	et.accsDup = types.MakeAccs("foo", "foo", "foo")

	et.store = types.NewMemKVStore()
	et.state = NewState(et.store)
	et.state.SetChainID(et.chainID)
}

//--------------------------------------------------------

func TestGetInputs(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//nil submissions
	acc, res := getInputs(nil, nil)
	assert.True(res.IsOK(), "getInputs: error on nil submission")
	assert.Zero(len(acc), "getInputs: accounts returned on nil submission")

	//test getInputs for registered, non-registered account
	et.reset()
	txs := types.Accs2TxInputs(et.accsFoo, 1)
	acc, res = getInputs(et.state, txs)
	assert.True(res.IsErr(), "getInputs: expected error when using getInput with non-registered Input")

	et.acc2State(et.accsFoo)
	acc, res = getInputs(et.state, txs)
	assert.True(res.IsOK(), "getInputs: expected to getInput from registered Input")

	//test sending duplicate accounts
	et.reset()
	et.acc2State(et.accsDup)
	txs = types.Accs2TxInputs(et.accsDup, 1)
	acc, res = getInputs(et.state, txs)
	assert.True(res.IsErr(), "getInputs: expected error when sending duplicate accounts")
}

func TestGetOrMakeOutputs(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//nil submissions
	acc, res := getOrMakeOutputs(nil, nil, nil)
	assert.True(res.IsOK(), "getOrMakeOutputs: error on nil submission")
	assert.Zero(len(acc), "getOrMakeOutputs: accounts returned on nil submission")

	//test sending duplicate accounts
	et.reset()
	txs := types.Accs2TxOutputs(et.accsDup)
	_, res = getOrMakeOutputs(et.state, nil, txs)
	assert.True(res.IsErr(), "getOrMakeOutputs: expected error when sending duplicate accounts")

	//test sending to existing/new account account
	et.reset()
	txs1 := types.Accs2TxOutputs(et.accsFoo)
	txs2 := types.Accs2TxOutputs(et.accsBar)

	et.acc2State(et.accsFoo)
	_, res = getOrMakeOutputs(et.state, nil, txs1)
	assert.True(res.IsOK(), "getOrMakeOutputs: error when sending to existing account")

	mapRes2, res := getOrMakeOutputs(et.state, nil, txs2)
	assert.True(res.IsOK(), "getOrMakeOutputs: error when sending to new account")

	//test the map results
	_, map2ok := mapRes2[string(txs2[0].Address)]
	assert.True(map2ok, "getOrMakeOutputs: account output does not contain new account map item")

}

func TestValidateInputsBasic(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input basic
	txs := types.Accs2TxInputs(et.accsFoo, 1)
	res := validateInputsBasic(txs)
	assert.True(res.IsOK(), fmt.Sprintf("validateInputsBasic: expected no error on good tx input. Error: %v", res.Error()))

	txs[0].Coins[0].Amount = 0
	res = validateInputsBasic(txs)
	assert.True(res.IsErr(), "validateInputsBasic: expected error on bad tx input")

}

func TestValidateInputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate inputs advanced
	txs := et.getTx(1, et.accsFooBar)

	et.acc2State(et.accsFooBar)
	accMap, res := getInputs(et.state, txs.Inputs)
	assert.True(res.IsOK(), fmt.Sprintf("validateInputsAdvanced: error retrieving accMap. Error: %v", res.Error()))
	signBytes := txs.SignBytes(et.chainID)

	//test bad case, unsigned
	totalCoins, res := validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.True(res.IsErr(), "validateInputsAdvanced: expected an error on an unsigned tx input")

	//test good case sgined
	et.signTx(txs, et.accsFooBar)
	totalCoins, res = validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.True(res.IsOK(), fmt.Sprintf("validateInputsAdvanced: expected no error on good tx input. Error: %v", res.Error()))
	assert.True(totalCoins.IsEqual(txs.Inputs[0].Coins.Plus(txs.Inputs[1].Coins)), "ValidateInputsAdvanced: transaction total coins are not equal")
}

func TestValidateInputAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input advanced
	txs := et.getTx(1, et.accsFooBar)

	et.acc2State(et.accsFooBar)
	signBytes := txs.SignBytes(et.chainID)

	//unsigned case
	res := validateInputAdvanced(&et.accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	et.signTx(txs, et.accsFooBar)
	res = validateInputAdvanced(&et.accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsOK(), fmt.Sprintf("validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error()))

	//bad sequence case
	et.accsFooBar[0].Sequence = 2
	et.signTx(txs, et.accsFooBar)
	res = validateInputAdvanced(&et.accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with bad sequence")
	et.accsFooBar[0].Account.Sequence = 1 //restore sequence

	//bad balance case
	et.accsFooBar[1].Balance = types.Coins{{"mycoin", 2}}
	et.signTx(txs, et.accsFooBar)
	res = validateInputAdvanced(&et.accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with insufficient funds")
}

func TestValidateOutputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validateOutputsBasic
	txs := types.Accs2TxOutputs(et.accsFoo)
	res := validateOutputsBasic(txs)
	assert.True(res.IsOK(), fmt.Sprintf("validateOutputsBasic: expected no error on good tx input. Error: %v", res.Error()))

	txs[0].Coins[0].Amount = 0
	res = validateOutputsBasic(txs)
	assert.True(res.IsErr(), fmt.Sprintf("validateInputBasic: expected error on bad tx inputi. Error: %v", res.Error()))
}

func TestSumOutput(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//SumOutput
	txs := types.Accs2TxOutputs(et.accsFooBar)
	total := sumOutputs(txs)
	assert.True(total.IsEqual(txs[0].Coins.Plus(txs[1].Coins)), "sumOutputs: total coins are not equal")
}

func TestAdjustBy(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//adjustByInputs/adjustByOutputs
	//sending transaction from Foo to Bar
	initBalFoo := et.accsFooBar[0].Account.Balance
	initBalBar := et.accsFooBar[1].Account.Balance
	et.acc2State(et.accsFooBar)

	txIn := types.Accs2TxInputs(et.accsFoo, 1)
	txOut := types.Accs2TxOutputs(et.accsBar)
	accMap, _ := getInputs(et.state, txIn)
	accMap, _ = getOrMakeOutputs(et.state, accMap, txOut)

	adjustByInputs(et.state, accMap, txIn)
	adjustByOutputs(et.state, accMap, txOut, false)

	endBalFoo := accMap[string(et.accsFooBar[0].Account.PubKey.Address())].Balance
	endBalBar := accMap[string(et.accsFooBar[1].Account.PubKey.Address())].Balance
	decrBalFoo := initBalFoo.Minus(endBalFoo)
	incrBalBar := endBalBar.Minus(initBalBar)

	assert.True(decrBalFoo.IsEqual(txIn[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalFoo.String(), txIn[0].Coins.String()))
	assert.True(incrBalBar.IsEqual(txOut[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalBar.String(), txOut[0].Coins.String()))

}

func TestExecTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//ExecTx
	txs := et.getTx(1, et.accsFoo)
	et.acc2State(et.accsFoo)
	et.acc2State(et.accsBar)
	et.signTx(txs, et.accsFoo)

	//Bad Balance
	et.accsFoo[0].Balance = types.Coins{{"mycoin", 2}}
	et.acc2State(et.accsFoo)
	res, _, _, _, _ := et.exec(txs, true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, foo, fooexp, bar, barexp := et.exec(txs, false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.False(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, foo: %v, fooExp: %v", foo, fooexp))
	assert.False(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, bar: %v, barExp: %v", bar, barexp))

	//Regular CheckTx
	et.reset()
	et.acc2State(et.accsFoo)
	et.acc2State(et.accsBar)
	res, _, _, _, _ = et.exec(txs, true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	et.reset()
	et.acc2State(et.accsFoo)
	et.acc2State(et.accsBar)
	res, foo, fooexp, bar, barexp = et.exec(txs, false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, bar: %v, barExp: %v", bar, barexp))
}
