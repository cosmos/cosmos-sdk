package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-crypto"
)

//States and Stores for tests
var (
	store                                 types.KVStore
	state                                 *State
	accsFoo, accsBar, accsFooBar, accsDup []types.PrivAccount
	chainID                               string = "test_chain_id"
)

func makeAccs(secrets []string) (accs []types.PrivAccount) {

	for _, secret := range secrets {
		privAcc := types.PrivAccountFromSecret(secret)
		privAcc.Account.Balance = types.Coins{{"mycoin", 1000}}
		accs = append(accs, privAcc)
	}
	return accs
}

func acc2State(accs []types.PrivAccount) {
	for _, acc := range accs {
		state.SetAccount(acc.Account.PubKey.Address(), &acc.Account)
	}
}

//each tx input signs the tx bytes
func signSend(tx *types.SendTx, accs []types.PrivAccount) {
	signBytes := tx.SignBytes(chainID)
	for i, _ := range tx.Inputs {
		tx.Inputs[i].Signature = crypto.SignatureS{accs[i].Sign(signBytes)}
	}
}

//turn a list of accounts into basic list of transaction inputs
func accs2TxInputs(accs []types.PrivAccount) []types.TxInput {
	var txs []types.TxInput
	for _, acc := range accs {
		tx := types.NewTxInput(
			acc.Account.PubKey,
			types.Coins{{"mycoin", 5}},
			1)
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

//reset the store/state/Inputs
func reset() {
	accsFoo = makeAccs([]string{"foo"})
	accsBar = makeAccs([]string{"bar"})
	accsFooBar = makeAccs([]string{"foo", "bar"})
	accsDup = makeAccs([]string{"foo", "foo", "foo"})

	store = types.NewMemKVStore()
	state = NewState(store)
	state.SetChainID(chainID)
}

func TestGetInputs(t *testing.T) {
	assert := assert.New(t)

	//nil submissions
	reset()
	acc, res := getInputs(nil, nil)
	assert.False(res.IsErr(), "getInputs: error on nil submission")
	assert.Zero(len(acc), "getInputs: accounts returned on nil submission")

	//test getInputs for registered, non-registered account
	reset()
	txs := accs2TxInputs(accsFoo)
	_, res = getInputs(state, txs)
	assert.True(res.IsErr(), "getInputs: expected to getInput from registered Input")

	acc2State(accsFoo)
	_, res = getInputs(state, txs)
	assert.False(res.IsErr(), "getInputs: expected to getInput from registered Input")

	//test sending duplicate accounts
	reset()
	acc2State(accsDup)
	txs = accs2TxInputs(accsDup)
	_, res = getInputs(state, txs)
	assert.True(res.IsErr(), "getInputs: expected error when sending duplicate accounts")
}

func TestGetOrMakeOutputs(t *testing.T) {
	assert := assert.New(t)

	//nil submissions
	reset()
	acc, res := getOrMakeOutputs(nil, nil, nil)
	assert.False(res.IsErr(), "getOrMakeOutputs: error on nil submission")
	assert.Zero(len(acc), "getOrMakeOutputs: accounts returned on nil submission")

	//test sending duplicate accounts
	reset()
	txs := accs2TxOutputs(accsDup)
	_, res = getOrMakeOutputs(state, nil, txs)
	assert.True(res.IsErr(), "getOrMakeOutputs: expected error when sending duplicate accounts")

	//test sending to existing/new account account
	reset()
	txs1 := accs2TxOutputs(accsFoo)
	txs2 := accs2TxOutputs(accsBar)

	acc2State(accsFoo)
	_, res = getOrMakeOutputs(state, nil, txs1)
	assert.False(res.IsErr(), "getOrMakeOutputs: error when sending to existing account")

	mapRes2, res := getOrMakeOutputs(state, nil, txs2)
	assert.False(res.IsErr(), "getOrMakeOutputs: error when sending to new account")

	//test the map results
	_, map2ok := mapRes2[string(txs2[0].Address)]
	assert.True(map2ok, "getOrMakeOutputs: account output does not contain new account map item")

}

func TestValidateInputsBasic(t *testing.T) {
	assert := assert.New(t)

	//validate input basic
	reset()
	txs := accs2TxInputs(accsFoo)
	res := validateInputsBasic(txs)
	assert.False(res.IsErr(), fmt.Sprintf("validateInputsBasic: expected no error on good tx input. Error: %v", res.Error()))

	txs[0].Coins[0].Amount = 0
	res = validateInputsBasic(txs)
	assert.True(res.IsErr(), "validateInputsBasic: expected error on bad tx input")

}

func TestValidateInputsAdvanced(t *testing.T) {
	assert := assert.New(t)
	//validate inputs advanced
	reset()
	txs := types.SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Inputs:  accs2TxInputs(accsFooBar),
		Outputs: accs2TxOutputs(accsBar),
	}

	acc2State(accsFooBar)
	accMap, res := getInputs(state, txs.Inputs)
	assert.False(res.IsErr(), fmt.Sprintf("validateInputsAdvanced: error retrieving accMap. Error: %v", res.Error()))
	signBytes := txs.SignBytes(chainID)

	//test bad case, unsigned
	totalCoins, res := validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.True(res.IsErr(), "validateInputsAdvanced: expected an error on an unsigned tx input")

	//test good case sgined
	signSend(&txs, accsFooBar)
	totalCoins, res = validateInputsAdvanced(accMap, signBytes, txs.Inputs)
	assert.False(res.IsErr(), fmt.Sprintf("validateInputsAdvanced: expected no error on good tx input. Error: %v", res.Error()))
	assert.True(totalCoins.IsEqual(txs.Inputs[0].Coins.Plus(txs.Inputs[1].Coins)), "ValidateInputsAdvanced: transaction total coins are not equal")
}

func TestValidateInputAdvanced(t *testing.T) {
	assert := assert.New(t)

	//validate input advanced
	reset()
	txs := types.SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Inputs:  accs2TxInputs(accsFooBar),
		Outputs: accs2TxOutputs(accsBar),
	}

	acc2State(accsFooBar)
	signBytes := txs.SignBytes(chainID)

	//unsigned case
	res := validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input without signature")

	//good signed case
	signSend(&txs, accsFooBar)
	res = validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.False(res.IsErr(), fmt.Sprintf("validateInputAdvanced: expected no error on good tx input. Error: %v", res.Error()))

	//bad sequence case
	accsFooBar[0].Sequence = 2
	signSend(&txs, accsFooBar)
	res = validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with bad sequence")
	accsFooBar[0].Account.Sequence = 1 //restore sequence

	//bad balance case
	accsFooBar[1].Balance = types.Coins{{"mycoin", 2}}
	signSend(&txs, accsFooBar)
	res = validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])
	assert.True(res.IsErr(), "validateInputAdvanced: expected error on tx input with insufficient funds")
}

func TestValidateOutputsAdvanced(t *testing.T) {
	assert := assert.New(t)

	//validateOutputsBasic
	reset()
	txs := accs2TxOutputs(accsFoo)
	res := validateOutputsBasic(txs)
	assert.False(res.IsErr(), fmt.Sprintf("validateOutputsBasic: expected no error on good tx input. Error: %v", res.Error()))

	txs[0].Coins[0].Amount = 0
	res = validateOutputsBasic(txs)
	assert.True(res.IsErr(), fmt.Sprintf("validateInputBasic: expected error on bad tx inputi. Error: %v", res.Error()))
}

func TestSumOutput(t *testing.T) {
	assert := assert.New(t)

	//SumOutput
	reset()
	txs := accs2TxOutputs(accsFooBar)
	total := sumOutputs(txs)
	assert.True(total.IsEqual(txs[0].Coins.Plus(txs[1].Coins)), "sumOutputs: total coins are not equal")
}

func TestAdjustBy(t *testing.T) {
	assert := assert.New(t)

	//adjustByInputs/adjustByOutputs
	//sending transaction from Foo to Bar
	reset()
	initBalFoo := accsFooBar[0].Account.Balance
	initBalBar := accsFooBar[1].Account.Balance
	acc2State(accsFooBar)

	txIn := accs2TxInputs(accsFoo)
	txOut := accs2TxOutputs(accsBar)
	accMap, _ := getInputs(state, txIn)
	accMap, _ = getOrMakeOutputs(state, accMap, txOut)

	adjustByInputs(state, accMap, txIn)
	adjustByOutputs(state, accMap, txOut, false)

	endBalFoo := accMap[string(accsFooBar[0].Account.PubKey.Address())].Balance
	endBalBar := accMap[string(accsFooBar[1].Account.PubKey.Address())].Balance
	decrBalFoo := initBalFoo.Minus(endBalFoo)
	incrBalBar := endBalBar.Minus(initBalBar)

	assert.True(decrBalFoo.IsEqual(txIn[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalFoo.String(), txIn[0].Coins.String()))
	assert.True(incrBalBar.IsEqual(txOut[0].Coins),
		fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalBar.String(), txOut[0].Coins.String()))

}

func TestExecTx(t *testing.T) {
	assert := assert.New(t)

	//ExecTx
	reset()
	txs := &types.SendTx{
		Gas:     0,
		Fee:     types.Coin{"mycoin", 1},
		Inputs:  accs2TxInputs(accsFoo),
		Outputs: accs2TxOutputs(accsBar),
	}

	acc2State(accsFoo)
	acc2State(accsBar)
	signSend(txs, accsFoo)

	exec := func(checkTx bool) (ExecTxRes abci.Result, foo, fooExp, bar, barExp types.Coins) {
		initBalFoo := state.GetAccount(accsFoo[0].Account.PubKey.Address()).Balance
		initBalBar := state.GetAccount(accsBar[0].Account.PubKey.Address()).Balance
		res := ExecTx(state, nil, txs, checkTx, nil)
		endBalFoo := state.GetAccount(accsFoo[0].Account.PubKey.Address()).Balance
		endBalBar := state.GetAccount(accsBar[0].Account.PubKey.Address()).Balance
		decrBalFooExp := txs.Outputs[0].Coins.Plus(types.Coins{txs.Fee})
		return res, endBalFoo, initBalFoo.Minus(decrBalFooExp), endBalBar, initBalBar.Plus(txs.Outputs[0].Coins)
	}

	//Bad Balance
	accsFoo[0].Balance = types.Coins{{"mycoin", 2}}
	acc2State(accsFoo)
	res, _, _, _, _ := exec(true)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad CheckTx: Expected error return from ExecTx, returned: %v", res))
	res, foo, fooexp, bar, barexp := exec(false)
	assert.True(res.IsErr(), fmt.Sprintf("ExecTx/Bad DeliverTx: Expected error return from ExecTx, returned: %v", res))
	assert.False(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, foo: %v, fooExp: %v", foo, fooexp))
	assert.False(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/Bad DeliverTx: shouldn't be equal, bar: %v, barExp: %v", bar, barexp))

	//Regular CheckTx
	reset()
	acc2State(accsFoo)
	acc2State(accsBar)
	res, _, _, _, _ = exec(true)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good CheckTx: Expected OK return from ExecTx, Error: %v", res))

	//Regular DeliverTx
	reset()
	acc2State(accsFoo)
	acc2State(accsBar)
	res, foo, fooexp, bar, barexp = exec(false)
	assert.True(res.IsOK(), fmt.Sprintf("ExecTx/Good DeliverTx: Expected OK return from ExecTx, Error: %v", res))
	assert.True(foo.IsEqual(fooexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in input coins, foo: %v, fooExp: %v", foo, fooexp))
	assert.True(bar.IsEqual(barexp), fmt.Sprintf("ExecTx/good DeliverTx: unexpected change in output coins, bar: %v, barExp: %v", bar, barexp))

}
