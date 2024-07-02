//go:build system_test

package systemtests

import (
	"strconv"
	"testing"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	// scenario: test unordered tx duplicate
	// given a running chain with a tx in the unordered tx pool
	// when a new tx with the same hash is broadcasted
	// then the new tx should be rejected

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	sut.StartChain(t)

	height := sut.CurrentHeight()
	timeoutHeight := height + 15
	timeoutHeightStr := strconv.Itoa(int(timeoutHeight))
	// send tokens
	rsp1 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--unordered=true", "--timeout-height="+timeoutHeightStr)
	rsp2 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--unordered=true", "--timeout-height="+timeoutHeightStr)
	t.Logf("rsp2: %s\n\n", rsp2)

	RequireTxSuccess(t, rsp1)
	RequireTxFailure(t, rsp2)
	// assert TX executed before timeout
	for cli.QueryBalance(account2Addr, "stake") != 5000 {
		t.Log("query balance")
		if current := sut.AwaitNextBlock(t); current > timeoutHeight {
			t.Fatal("TX was not executed before timeout")
		}
	}
}
