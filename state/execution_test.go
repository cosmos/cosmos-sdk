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
	chainID string
	store   types.KVStore
	state   *State
	accIn   types.PrivAccount
	accOut  types.PrivAccount
}

func newExecTest() *execTest {
	et := &execTest{
		chainID: "test_chain_id",
	}
	et.reset()
	return et
}

func (et *execTest) signTx(tx *types.SendTx, accsIn ...types.PrivAccount) {
	types.SignTx(et.chainID, tx, accsIn...)
}

// make tx from accsIn to et.accOut
func (et *execTest) getTx(seq int, accsIn ...types.PrivAccount) *types.SendTx {
	return types.GetTx(seq, et.accOut, accsIn...)
}

// returns the final balance and expected balance for input and output accounts
func (et *execTest) exec(tx *types.SendTx, checkTx bool) (res abci.Result, inGot, inExp, outGot, outExp types.Coins) {
	initBalIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	initBalOut := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance

	res = ExecTx(et.state, nil, tx, checkTx, nil)

	endBalIn := et.state.GetAccount(et.accIn.Account.PubKey.Address()).Balance
	endBalOut := et.state.GetAccount(et.accOut.Account.PubKey.Address()).Balance
	decrBalInExp := tx.Outputs[0].Coins.Plus(types.Coins{tx.Fee}) //expected decrease in balance In
	return res, endBalIn, initBalIn.Minus(decrBalInExp), endBalOut, initBalOut.Plus(tx.Outputs[0].Coins)
}

func (et *execTest) acc2State(accs ...types.PrivAccount) {
	for _, acc := range accs {
		et.state.SetAccount(acc.Account.PubKey.Address(), &acc.Account)
	}
}

//reset everything. state is empty
func (et *execTest) reset() {
	et.accIn = types.MakeAcc("foo")
	et.accOut = types.MakeAcc("bar")

	et.store = types.NewMemKVStore()
	et.state = NewState(et.store)
	et.state.SetChainID(et.chainID)

	// NOTE we dont run acc2State here
	// so we can test non-existing accounts

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
	txs := types.Accs2TxInputs(1, et.accIn)
	acc, res = getInputs(et.state, txs)
	assert.True(res.IsErr(), "getInputs: expected error when using getInput with non-registered Input")

	et.acc2State(et.accIn)
	acc, res = getInputs(et.state, txs)
	assert.True(res.IsOK(), "getInputs: expected to getInput from registered Input")

	//test sending duplicate accounts
	et.reset()
	et.acc2State(et.accIn, et.accIn, et.accIn)
	txs = types.Accs2TxInputs(1, et.accIn, et.accIn, et.accIn)
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
	txs := types.Accs2TxOutputs(et.accIn, et.accIn, et.accIn)
	_, res = getOrMakeOutputs(et.state, nil, txs)
	assert.True(res.IsErr(), "getOrMakeOutputs: expected error when sending duplicate accounts")

	//test sending to existing/new account
	et.reset()
	txs1 := types.Accs2TxOutputs(et.accIn)
	txs2 := types.Accs2TxOutputs(et.accOut)

	et.acc2State(et.accIn)
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
	txs := types.Accs2TxInputs(1, et.accIn)
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
	txs := et.getTx(1, et.accIn, et.accOut)

	et.acc2State(et.accIn, et.accOut)
	accMap, res := getInputs(et.state, txs.Inputs)
	assert.True(res.IsOK(), fmt.Sprintf("validateInputsAdvanced: error retrieving accMap. Error: %v", res.Error()))
	signBytes := txs.SignBytes(et.chainID)

	//test bad case, unsigned
	totalCoins, res := validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.True(res.IsErr(), "validateInputsAdvanced: expected an error on an unsigned tx input")

	//test good case sgined
	et.signTx(txs, et.accIn, et.accOut)
	totalCoins, res = validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.True(res.IsOK(), fmt.Sprintf("validateInputsAdvanced: expected no error on good tx input. Error: %v", res.Error()))
	assert.True(totalCoins.IsEqual(txs.Inputs[0].Coins.Plus(txs.Inputs[1].Coins)), "ValidateInputsAdvanced: transaction total coins are not equal")
}

func TestValidateInputAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validate input advanced
	txs := et.getTx(1, et.accIn, et.accOut)

	et.acc2State(et.accIn, et.accOut)
	signBytes := txs.SignBytes(et.chainID)

	//unsigned case
	res := validateInputAdvanced(&et.accIn.Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	et.signTx(txs, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, txs.Inputs[0])
	assert.True(res.IsOK(), fmt.Sprintf("validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error()))

	//bad sequence case
	et.accIn.Sequence = 2
	et.signTx(txs, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with bad sequence")
	et.accIn.Sequence = 1 //restore sequence

	//bad balance case
	et.accOut.Balance = types.Coins{{"mycoin", 2}}
	et.signTx(txs, et.accIn, et.accOut)
	res = validateInputAdvanced(&et.accIn.Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with insufficient funds")
}

func TestValidateOutputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//validateOutputsBasic
	txs := types.Accs2TxOutputs(et.accIn)
	res := validateOutputsBasic(txs)
	assert.True(res.IsOK(), fmt.Sprintf("validateOutputsBasic: expected no error on good tx output. Error: %v", res.Error()))

	txs[0].Coins[0].Amount = 0
	res = validateOutputsBasic(txs)
	assert.True(res.IsErr(), fmt.Sprintf("validateInputBasic: expected error on bad tx output. Error: %v", res.Error()))
}

func TestSumOutput(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//SumOutput
	txs := types.Accs2TxOutputs(et.accIn, et.accOut)
	total := sumOutputs(txs)
	assert.True(total.IsEqual(txs[0].Coins.Plus(txs[1].Coins)), "sumOutputs: total coins are not equal")
}

func TestAdjustBy(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//adjustByInputs/adjustByOutputs
	//sending transaction from accIn to accOut
	initBalIn := et.accIn.Account.Balance
	initBalOut := et.accOut.Account.Balance
	et.acc2State(et.accIn, et.accOut)

	txIn := types.Accs2TxInputs(1, et.accIn)
	txOut := types.Accs2TxOutputs(et.accOut)
	accMap, _ := getInputs(et.state, txIn)
	accMap, _ = getOrMakeOutputs(et.state, accMap, txOut)

	adjustByInputs(et.state, accMap, txIn)
	adjustByOutputs(et.state, accMap, txOut, false)

	endBalIn := accMap[string(et.accIn.Account.PubKey.Address())].Balance
	endBalOut := accMap[string(et.accOut.Account.PubKey.Address())].Balance
	decrBalIn := initBalIn.Minus(endBalIn)
	incrBalOut := endBalOut.Minus(initBalOut)

	assert.True(decrBalIn.IsEqual(txIn[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalIn.String(), txIn[0].Coins.String()))
	assert.True(incrBalOut.IsEqual(txOut[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalOut.String(), txOut[0].Coins.String()))

}

func TestExecTx(t *testing.T) {
	assert := assert.New(t)
	et := newExecTest()

	//ExecTx
	txs := et.getTx(1, et.accIn)
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	et.signTx(txs, et.accIn)

	//Bad Balance
	et.accIn.Balance = types.Coins{{"mycoin", 2}}
	et.acc2State(et.accIn)
	res, _, _, _, _ := et.exec(txs, true)
	assert.True(res.IsErr(),
		fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))

	res, balIn, balInExp, balOut, balOutExp := et.exec(txs, false)
	assert.True(res.IsErr(),
		fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.False(balIn.IsEqual(balInExp),
		fmt.Sprintf("ExecTx/Bad DeliverTx: balance shouldn't be equal for accIn: got %v, expected: %v", balIn, balInExp))
	assert.False(balOut.IsEqual(balOutExp),
		fmt.Sprintf("ExecTx/Bad DeliverTx: balance shouldn't be equal for accOut: got %v, expected: %v", balOut, balOutExp))

	//Regular CheckTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, _, _, _, _ = et.exec(txs, true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	et.reset()
	et.acc2State(et.accIn)
	et.acc2State(et.accOut)
	res, balIn, balInExp, balOut, balOutExp = et.exec(txs, false)
	assert.True(res.IsOK(),
		fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(balIn.IsEqual(balInExp),
		fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input balance, got: %v, expected: %v", balIn, balInExp))
	assert.True(balOut.IsEqual(balOutExp),
		fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output balance, got: %v, expected: %v", balOut, balOutExp))
}
