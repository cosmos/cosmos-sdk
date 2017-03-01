package state

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-crypto"
)

func TestExecution(t *testing.T) {

	//States and Stores for tests
	var store types.KVStore
	var state *State
	var accsFoo, accsBar, accsFooBar, accsDup []types.PrivAccount
	chainID := "test_chain_id"

	makeAccs := func(secrets []string) (accs []types.PrivAccount) {

		for _, secret := range secrets {
			privAcc := types.PrivAccountFromSecret(secret)
			privAcc.Account.Balance = types.Coins{{"mycoin", 1000}}
			accs = append(accs, privAcc)
		}
		return accs
	}

	acc2State := func(accs []types.PrivAccount) {
		for _, acc := range accs {
			state.SetAccount(acc.Account.PubKey.Address(), &acc.Account)
		}
	}

	//each tx input signs the tx bytes
	signSend := func(tx *types.SendTx, accs []types.PrivAccount) {
		signBytes := tx.SignBytes(chainID)
		for i, _ := range tx.Inputs {
			tx.Inputs[i].Signature = crypto.SignatureS{accs[i].Sign(signBytes)}
		}
	}

	//turn a list of accounts into basic list of transaction inputs
	accs2TxInputs := func(accs []types.PrivAccount) []types.TxInput {
		var txs []types.TxInput
		for _, acc := range accs {
			tx := types.NewTxInput(
				acc.Account.PubKey,
				types.Coins{{"mycoin", 10}},
				1)
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
				types.Coins{{"mycoin", 9}}}
			txs = append(txs, tx)
		}
		return txs
	}

	//reset the store/state/Inputs
	reset := func() {
		accsFoo = makeAccs([]string{"foo"})
		accsBar = makeAccs([]string{"bar"})
		accsFooBar = makeAccs([]string{"foo", "bar"})
		accsDup = makeAccs([]string{"foo", "foo", "foo"})

		store = types.NewMemKVStore()
		state = NewState(store)
		state.SetChainID(chainID)
	}

	type er struct {
		exp bool   //assert true
		msg string //msg is assert fails
	}

	//define the test list
	testList := []struct {
		tester func() []er
	}{
		///////////////
		//getInputs

		//nil submissions
		{func() []er {
			acc, res := getInputs(nil, nil)
			return []er{
				{!res.IsErr(), "getInputs: error on nil submission"},
				{len(acc) == 0, "getInputs: accounts returned on nil submission"},
			}
		}},

		//test getInputs for registered, non-registered account
		{func() []er {
			txs := accs2TxInputs(accsFoo)
			_, res1 := getInputs(state, txs)
			acc2State(accsFoo)
			_, res2 := getInputs(state, txs)
			return []er{
				{res1.IsErr(), "getInputs: expected to getInput from registered Input"},
				{!res2.IsErr(), "getInputs: expected to getInput from registered Input"},
			}
		}},

		//test sending duplicate accounts
		{func() []er {
			acc2State(accsDup)
			txs := accs2TxInputs(accsDup)
			_, res := getInputs(state, txs)
			return []er{{res.IsErr(), "getInputs: expected error when sending duplicate accounts"}}
		}},

		///////////////////
		//getOrMakeOutputs

		//nil submissions
		{func() []er {
			acc, res := getOrMakeOutputs(nil, nil, nil)
			return []er{
				{!res.IsErr(), "getOrMakeOutputs: error on nil submission"},
				{len(acc) == 0, "getOrMakeOutputs: accounts returned on nil submission"},
			}
		}},

		//test sending duplicate accounts
		{func() []er {
			txs := accs2TxOutputs(accsDup)
			_, res := getOrMakeOutputs(state, nil, txs)
			return []er{{res.IsErr(), "getOrMakeOutputs: expected error when sending duplicate accounts"}}
		}},

		//test sending to existing/new account account
		{func() []er {
			txs1 := accs2TxOutputs(accsFoo)
			txs2 := accs2TxOutputs(accsBar)

			acc2State(accsFoo)
			_, res1 := getOrMakeOutputs(state, nil, txs1)
			mapRes2, res2 := getOrMakeOutputs(state, nil, txs2)

			//TODO Fix this commented out test
			//test the map results
			//acc2, map2ok := mapRes2[string(txs2[0].Address)]
			_, map2ok := mapRes2[string(txs2[0].Address)]

			return []er{
				{!res1.IsErr(), "getOrMakeOutputs: error when sending to existing account"},
				{!res2.IsErr(), "getOrMakeOutputs: error when sending to new account"},
				{map2ok, "getOrMakeOutputs: account output does not contain new account map item"},
			}
			//{accs2[0].PubKey.Equals(acc2.PubKey), "getOrMakeOutputs: account output does not contain new account pointer"}}
		}},

		//validate input basic
		{func() []er {
			txs := accs2TxInputs(accsFoo)
			res1 := validateInputsBasic(txs)
			txs[0].Coins[0].Amount = 0
			res2 := validateInputsBasic(txs)
			return []er{
				{!res1.IsErr(), fmt.Sprintf("validateInputsBasic: expected no error on good tx input. Error: %v", res1.Error())},
				{res2.IsErr(), "validateInputsBasic: expected error on bad tx input"},
			}
		}},

		//validate inputs advanced
		{func() []er {
			txs := types.SendTx{
				Gas:     0,
				Fee:     types.Coin{"mycoin", 1},
				Inputs:  accs2TxInputs(accsFooBar),
				Outputs: accs2TxOutputs(accsBar),
			}

			acc2State(accsFooBar)
			accMap, res1 := getInputs(state, txs.Inputs)
			signBytes := txs.SignBytes(chainID)

			//test bad case, unsigned
			totalCoins, res2 := validateInputsAdvanced(accMap, signBytes, txs.Inputs)

			//test good case sgined
			signSend(&txs, accsFooBar)
			totalCoins, res3 := validateInputsAdvanced(accMap, signBytes, txs.Inputs)

			return []er{
				{!res1.IsErr(), fmt.Sprintf("validateInputsAdvanced: error retrieving accMap. Error: %v", res1.Error())},
				{res2.IsErr(), "validateInputsAdvanced: expected an error on an unsigned tx input"},
				{!res3.IsErr(), fmt.Sprintf("validateInputsAdvanced: expected no error on good tx input. Error: %v", res3.Error())},
				{totalCoins.IsEqual(txs.Inputs[0].Coins.Plus(txs.Inputs[1].Coins)), "ValidateInputsAdvanced: transaction total coins are not equal"},
			}
		}},

		//validate input advanced
		{func() []er {
			txs := types.SendTx{
				Gas:     0,
				Fee:     types.Coin{"mycoin", 1},
				Inputs:  accs2TxInputs(accsFooBar),
				Outputs: accs2TxOutputs(accsBar),
			}

			acc2State(accsFooBar)
			signBytes := txs.SignBytes(chainID)

			//unsigned case
			res1 := validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])

			//good signed case
			signSend(&txs, accsFooBar)
			res2 := validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])

			//bad sequence case
			accsFooBar[0].Sequence = 2
			signSend(&txs, accsFooBar)
			res3 := validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])
			accsFooBar[0].Account.Sequence = 1 //restore sequence

			//bad balance case
			accsFooBar[1].Balance = types.Coins{{"mycoin", 2}}
			signSend(&txs, accsFooBar)
			res4 := validateInputAdvanced(&accsFooBar[0].Account, signBytes, txs.Inputs[0])

			return []er{
				{res1.IsErr(), "validateInputAdvanced: expected error on tx input without signature"},
				{!res2.IsErr(), fmt.Sprintf("validateInputAdvanced: expected no error on good tx input. Error: %v", res1.Error())},
				{res3.IsErr(), "validateInputAdvanced: expected error on tx input with bad sequence"},
				{res4.IsErr(), "validateInputAdvanced: expected error on tx input with insufficient funds"},
			}
		}},

		//validateOutputsBasic
		{func() []er {
			txs := accs2TxOutputs(accsFoo)
			res1 := validateOutputsBasic(txs)
			txs[0].Coins[0].Amount = 0
			res2 := validateOutputsBasic(txs)
			return []er{{!res1.IsErr(), fmt.Sprintf("validateOutputsBasic: expected no error on good tx input. Error: %v", res1.Error())},
				{res2.IsErr(), fmt.Sprintf("validateInputBasic: expected error on bad tx inputi. Error: %v", res2.Error())}}
		}},

		//SumOutput
		{func() []er {
			txs := accs2TxOutputs(accsFooBar)
			total := sumOutputs(txs)
			return []er{{total.IsEqual(txs[0].Coins.Plus(txs[1].Coins)), "sumOutputs: total coins are not equal"}}
		}},

		//adjustByInputs/adjustByOutputs
		//sending transaction from Foo to Bar
		{func() []er {

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

			return []er{
				{decrBalFoo.IsEqual(txIn[0].Coins),
					fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", decrBalFoo.String(), txIn[0].Coins.String())},
				{incrBalBar.IsEqual(txOut[0].Coins),
					fmt.Sprintf("adjustByInputs: total coins are not equal. diff: %v, tx: %v", incrBalBar.String(), txOut[0].Coins.String())},
			}
		}},

		//ExecTx
		{func() []er {
			txs := &types.SendTx{
				Gas:     0,
				Fee:     types.Coin{"mycoin", 1},
				Inputs:  accs2TxInputs(accsFoo),
				Outputs: accs2TxOutputs(accsBar),
			}

			initBalFoo := accsFooBar[0].Account.Balance
			initBalBar := accsFooBar[1].Account.Balance
			acc2State(accsFooBar)

			//sign that puppy
			signBytes := txs.SignBytes(chainID)
			sig := accsFoo[0].Sign(signBytes)
			txs.Inputs[0].Signature = crypto.SignatureS{sig}

			//TODO tests for CheckTx, some bad transactions
			err := ExecTx(state, nil, txs, false, nil)
			fmt.Println("ERR", err)

			endBalFoo := state.GetAccount(accsFooBar[0].Account.PubKey.Address()).Balance
			endBalBar := state.GetAccount(accsFooBar[1].Account.PubKey.Address()).Balance
			decrBalFoo := initBalFoo.Minus(endBalFoo)
			decrBalFooExp := txs.Outputs[0].Coins.Plus(types.Coins{txs.Fee})
			incrBalBar := endBalBar.Minus(initBalBar)

			return []er{
				{decrBalFoo.IsEqual(decrBalFooExp),
					fmt.Sprintf("ExecTx(sendTx): unexpected change in input coins. exp: %v, change: %v", decrBalFooExp.String(), decrBalFoo.String())},
				{incrBalBar.IsEqual(txs.Outputs[0].Coins),
					fmt.Sprintf("ExecTx(sendTx): unexpected change in output coins. exp: %v, change: %v", incrBalBar.String(), txs.Outputs[0].Coins.String())},
			}
		}},
	}

	//execute the tests
	for _, tl := range testList {
		reset()
		for _, tr := range tl.tester() { //loop through all outputs of a test
			assert.True(t, tr.exp, tr.msg)
		}
	}

}
