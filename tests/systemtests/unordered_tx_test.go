//go:build system_test

package systemtests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "github.com/cosmos/cosmos-sdk/testutil/systemtests"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	// scenario: test unordered tx duplicate
	// given a running chain with a tx in the unordered tx pool
	// when a new tx with the same hash is broadcasted
	// then the new tx should be rejected

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	timeoutTimestamp := time.Now().Add(time.Minute)
	// send tokens
	cmd := []string{"tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from=" + account1Addr, "--fees=1stake", fmt.Sprintf("--timeout-timestamp=%v", timeoutTimestamp.Unix()), "--unordered", "--sequence=1", "--note=1"}
	rsp1 := cli.Run(cmd...)
	systest.RequireTxSuccess(t, rsp1)

	assertDuplicateErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		output := gotOutputs[0].(string)
		code := gjson.Get(output, "code")
		require.True(t, code.Exists())
		require.Equal(t, int64(19), code.Int()) // 19 == already in mempool.
		return false                            // always abort
	}
	rsp2 := cli.WithRunErrorMatcher(assertDuplicateErr).Run(cmd...)
	systest.RequireTxFailure(t, rsp2)

	require.Eventually(t, func() bool {
		return cli.QueryBalance(account2Addr, "stake") == 5000
	}, 10*systest.Sut.BlockTime(), 200*time.Millisecond, "TX was not executed before timeout")
}
