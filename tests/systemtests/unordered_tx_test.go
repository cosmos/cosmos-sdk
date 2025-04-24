//go:build system_test

package systemtests

import (
	"fmt"
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

func TestTxBackwardsCompatability(t *testing.T) {
	// Scenario:
	// A transaction generated from a v0.53 chain without unordered and timeout_timestamp flags set should succeed.
	// Conversely, a transaction generated from a v0.53 chain with unordered and timeout_timestamp flags set should fail.
	var (
		denom                = "stake"
		transferAmount int64 = 1000
	)
	systest.Sut.ResetChain(t)

	v53CLI := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// we just get val addr for an address to send things to.
	valAddr := v53CLI.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// generate a deterministic account. we'll use this seed again later in the v50 chain.
	senderAddr := v53CLI.AddKeyFromSeed("account1", testSeed)

	v50CLI, legacySut := createLegacyBinary(t, initAccount{
		address: senderAddr,
		balance: "10000000000stake",
	})
	legacySut.StartChain(t)

	bankSendCmdArgs := []string{"tx", "bank", "send", senderAddr, valAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + v50CLI.ChainID(), "--fees=10stake", "--sign-mode=direct"}
	res, ok := v53CLI.RunOnly(bankSendCmdArgs...)
	require.True(t, ok)

	response, ok := v50CLI.AwaitTxCommitted(res, 15*time.Second)
	require.True(t, ok)
	code := gjson.Get(response, "code").Int()
	require.Equal(t, int64(0), code)

	bankSendCmdArgs = []string{"tx", "bank", "send", senderAddr, valAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + v50CLI.ChainID(), "--fees=10stake", "--sign-mode=direct", "--unordered", "--timeout-duration=8m"}
	res, ok = v53CLI.RunOnly(bankSendCmdArgs...)
	require.True(t, ok)

	code = gjson.Get(res, "code").Int()
	require.Equal(t, int64(2), code)
	require.Contains(t, res, "errUnknownField")
	legacySut.StopChain()

	// Now start a v53 chain, and send a transaction from a v50 client.
	// generate a deterministic account. we'll use this seed again later in the v50 chain.
	systest.Sut.SetupChain()
	systest.Sut.ModifyGenesisCLI(t,
		// we need our sender to be account 5 because that's how it was signed in the v53 scenario.
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000000stake"},
	)
	systest.Sut.StartChain(t)

	senderAddr = v50CLI.AddKeyFromSeed("account1", testSeed)
	bankSendCmdArgs = []string{"tx", "bank", "send", senderAddr, valAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + v50CLI.ChainID(), "--fees=10stake", "--sign-mode=direct"}
	res, ok = v50CLI.RunOnly(bankSendCmdArgs...)
	require.True(t, ok)

	response, ok = v53CLI.AwaitTxCommitted(res, 15*time.Second)
	require.True(t, ok)
	code = gjson.Get(response, "code").Int()
	require.Equal(t, int64(0), code)
}
