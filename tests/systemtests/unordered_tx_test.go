//go:build system_test

package systemtests

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
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

func TestTxBackwardsCompatability(t *testing.T) {
	// Scenario:
	// A transaction generated from a v0.53 chain without unordered and timeout_timestamp flags set should succeed.
	// Conversely, a transaction generated from a v0.53 chain with unordered and timeout_timestamp flags set should fail.
	var (
		denom                = "stake"
		transferAmount int64 = 1000
		testSeed             = "scene learn remember glide apple expand quality spawn property shoe lamp carry upset blossom draft reject aim file trash miss script joy only measure"
	)
	systest.Sut.ResetChain(t)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)
	senderAddr := cli.AddKeyFromSeed("account1", testSeed)
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000000stake"},
	)

	systest.Sut.StartChain(t)

	// create unsigned tx
	bankSendCmdArgs := []string{"tx", "bank", "send", senderAddr, valAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + cli.ChainID(), "--fees=10stake", "--sign-mode=direct", "--generate-only"}
	res := cli.RunCommandWithArgs(bankSendCmdArgs...)
	txFile := systest.StoreTempFile(t, []byte(res))
	res = cli.RunCommandWithArgs("tx", "sign", txFile.Name(), "--from="+senderAddr, "--chain-id="+cli.ChainID(), "--home="+systest.Sut.NodeDir(0))

	// fix the transaction body. we need to remove the fields. for now. this should be fixed later.
	var transaction map[string]any
	err := json.Unmarshal([]byte(res), &transaction)
	require.NoError(t, err)
	body := transaction["body"].(map[string]any)
	delete(body, "unordered")
	delete(body, "timeout_timestamp")
	transaction["body"] = body
	txBz, err := json.Marshal(transaction)
	require.NoError(t, err)
	signedTxFile := systest.StoreTempFile(t, txBz)
	systest.Sut.StopChain()

	//// Now we're going to switch to a v.50 chain.
	legacyBinary := systest.WorkDir + "/binaries/simdv50"

	legacySut := systest.NewSystemUnderTest(legacyBinary, systest.Verbose, 1, 1*time.Second)
	legacySut.SetTestnetInitializer(systest.LegacyInitializerWithBinary(legacyBinary, legacySut))
	legacySut.SetupChain()
	legacySut.SetExecBinary(legacyBinary)
	cli = systest.NewCLIWrapper(t, legacySut, systest.Verbose)
	legacySut.ModifyGenesisCLI(t,
		// add some bogus accounts because the v53 chain had 4 nodes which takes account numbers 1-4.
		[]string{"genesis", "add-genesis-account", cli.AddKey("foo"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", cli.AddKey("bar"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", cli.AddKey("baz"), "10000000000stake"},
		// we need our sender to be account 5 because thats how it was signed in the v53 scenario.
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000000stake"},
	)
	senderAddr = cli.AddKeyFromSeed("account1", testSeed)
	legacySut.StartChain(t)
	t.Logf("legacy sender address: %s", senderAddr)

	res = cli.Run("tx", "broadcast", signedTxFile.Name())
	systest.RequireTxSuccess(t, res)
}
