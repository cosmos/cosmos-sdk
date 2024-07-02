//go:build system_test

package systemtests

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	rsp1 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--timeout-height="+timeoutHeightStr, "--unordered", "--sequence=1", "--note=1")
	RequireTxSuccess(t, rsp1)

	assertDuplicateErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		assert.Contains(t, gotOutputs[0], "is duplicated: invalid request")
		return false // always abort
	}
	rsp2 := cli.WithRunErrorMatcher(assertDuplicateErr).Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", "--timeout-height="+timeoutHeightStr, "--unordered", "--sequence=1")
	RequireTxFailure(t, rsp2)

	// assert TX executed before timeout
	for cli.QueryBalance(account2Addr, "stake") != 5000 {
		t.Log("query balance")
		if current := sut.AwaitNextBlock(t); current > timeoutHeight {
			t.Fatal("TX was not executed before timeout")
		}
	}
}
