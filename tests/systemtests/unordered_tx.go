//go:build system_tests

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

	height := cli.QueryHeight()

	timeoutHeight := strconv.Itoa(int(height + 10))

	// send tokens
	rsp1 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--unordered=true", "--timeout-height="+timeoutHeight)
	rsp2 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--unordered=true", "--timeout-height="+timeoutHeight)
	RequireTxSuccess(t, rsp1)
	RequireTxFailure(t, rsp2)
}
