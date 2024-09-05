//go:build system_test

package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	t.Skip("The unordered tx antehanlder is missing in v2")
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

	timeoutTimestamp := time.Now().Add(time.Minute)
	// send tokens
	rsp1 := cli.Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", fmt.Sprintf("--timeout-timestamp=%v", timeoutTimestamp.Unix()), "--unordered", "--sequence=1", "--note=1")
	RequireTxSuccess(t, rsp1)

	assertDuplicateErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		assert.Contains(t, gotOutputs[0], "is duplicated: invalid request")
		return false // always abort
	}
	rsp2 := cli.WithRunErrorMatcher(assertDuplicateErr).Run("tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from="+account1Addr, "--fees=1stake", fmt.Sprintf("--timeout-timestamp=%v", timeoutTimestamp.Unix()), "--unordered", "--sequence=1")
	RequireTxFailure(t, rsp2)

	require.Eventually(t, func() bool {
		return cli.QueryBalance(account2Addr, "stake") == 5000
	}, time.Minute, time.Microsecond*500, "TX was not executed before timeout")
}
