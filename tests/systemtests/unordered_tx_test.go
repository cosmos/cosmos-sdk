//go:build system_test

package systemtests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	// scenario: test unordered tx duplicate
	// given a running chain with a tx in the unordered tx pool
	// when a new tx with the same unordered nonce is broadcasted,
	// then the new tx should be rejected.

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	// send tokens
	cmd := []string{"tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from=" + account1Addr, "--fees=1stake", "--timeout-duration=5m", "--unordered", "--note=1", "--chain-id=testing", "--generate-only"}
	rsp1 := cli.RunCommandWithArgs(cmd...)
	txFile := testutil.TempFile(t)
	_, err := txFile.WriteString(rsp1)
	require.NoError(t, err)

	signCmd := []string{"tx", "sign", txFile.Name(), "--from=" + account1Addr, "--chain-id=testing"}
	rsp1 = cli.RunCommandWithArgs(signCmd...)
	signedFile := testutil.TempFile(t)
	_, err = signedFile.WriteString(rsp1)
	require.NoError(t, err)

	cmd = []string{"tx", "broadcast", signedFile.Name(), "--chain-id=testing"}
	rsp1 = cli.RunCommandWithArgs(cmd...)
	systest.RequireTxSuccess(t, rsp1)

	cmd = []string{"tx", "broadcast", signedFile.Name(), "--chain-id=testing"}
	rsp2, _ := cli.RunOnly(cmd...)
	systest.RequireTxFailure(t, rsp2)
	code := gjson.Get(rsp2, "code")
	require.True(t, code.Exists())
	require.Equal(t, int64(19), code.Int())

	require.Eventually(t, func() bool {
		return cli.QueryBalance(account2Addr, "stake") == 5000
	}, 10*systest.Sut.BlockTime(), 200*time.Millisecond, "TX was not executed before timeout")
}
