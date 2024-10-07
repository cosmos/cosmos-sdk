package systemtests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const (
	authTestDenom = "stake"
)

func TestAuthSignAndBroadcastTxCmd(t *testing.T) {
	// scenario: test auth sign and broadcast commands
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	// get validator addresses
	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	sut.StartChain(t)

	val1Bal := cli.QueryBalance(val1Addr, authTestDenom)
	val2Bal := cli.QueryBalance(val2Addr, authTestDenom)
	var transferAmount int64 = 1000
	var feeAmount int64 = 1

	// run bank tx send with --generate-only flag
	bankSendGenCmd := []string{
		"tx", "bank", "send", val1Addr, val2Addr,
		fmt.Sprintf("%d%s", transferAmount, authTestDenom),
		fmt.Sprintf("--fees=%d%s", feeAmount, authTestDenom),
		"--generate-only",
	}

	var opFile *os.File

	assertGenOnlyOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		rsp := gotOutputs[0].(string)
		// get msg from output
		msgs := gjson.Get(rsp, "body.messages").Array()
		require.Len(t, msgs, 1)
		// check from address is equal to account1 address
		fromAddr := gjson.Get(msgs[0].String(), "from_address").String()
		require.Equal(t, val1Addr, fromAddr)
		// check to address is equal to account2 address
		toAddr := gjson.Get(msgs[0].String(), "to_address").String()
		require.Equal(t, val2Addr, toAddr)

		// write to temp file
		opFile = storeTempFile(t, []byte(rsp))
		defer opFile.Close()

		return false
	}
	_ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(bankSendGenCmd...)

	// query node0 account details
	rsp := cli.CustomQuery("q", "auth", "account", val1Addr)
	accSeq := gjson.Get(rsp, "account.value.sequence").Int()
	// as node0 is the first account, assume accNum is 0
	accNum := 0
	signTxCmd := []string{"tx", "sign", opFile.Name(), "--from=" + val1Addr}

	testCases := []struct {
		name      string
		extraArgs []string
	}{
		{
			"valid tx sign with offline mode",
			[]string{
				"--offline",
				fmt.Sprintf("--account-number=%d", accNum),
				fmt.Sprintf("--sequence=%d", accSeq),
			},
		},
		{
			"valid tx sign",
			[]string{},
		},
		{
			"valid tx sign with sign-mode",
			[]string{"--sign-mode=amino-json"},
		},
	}

	var signFile *os.File

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(signTxCmd, tc.extraArgs...)
			assertSignTxOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Len(t, gotOutputs, 1)
				rsp := gotOutputs[0].(string)

				signatures := gjson.Get(rsp, "signatures").Array()
				require.Len(t, signatures, 1)

				// write to temp file
				signFile = storeTempFile(t, []byte(rsp))
				defer signFile.Close()

				return false
			}

			// run tx sign command
			_ = cli.WithRunErrorMatcher(assertSignTxOutput).Run(cmd...)

			// validate signature
			rsp = cli.RunCommandWithArgs(cli.withTXFlags("tx", "validate-signatures", signFile.Name(), "--from="+val1Addr)...)
			require.Contains(t, rsp, val1Addr)

			// run broadcast tx command
			broadcastCmd := []string{"tx", "broadcast", signFile.Name()}
			rsp = cli.RunAndWait(broadcastCmd...)
			RequireTxSuccess(t, rsp)

			// query balance and confirm transaction
			expVal1Bal := val1Bal - transferAmount - feeAmount
			val1Bal = cli.QueryBalance(val1Addr, authTestDenom)
			require.Equal(t, expVal1Bal, val1Bal)

			expVal2Bal := val2Bal + transferAmount
			val2Bal = cli.QueryBalance(val2Addr, authTestDenom)
			require.Equal(t, expVal2Bal, val2Bal)
		})
	}

	// test broadcast with empty public key in signed tx
	rsp = cli.RunCommandWithArgs(cli.withTXFlags("tx", "sign", opFile.Name(), "--from="+val1Addr)...)
	updated, err := sjson.Set(rsp, "auth_info.signer_infos.0.public_key", nil)
	require.NoError(t, err)
	newSignFile := storeTempFile(t, []byte(updated))
	defer newSignFile.Close()

	broadcastCmd := []string{"tx", "broadcast", newSignFile.Name()}
	rsp = cli.RunCommandWithArgs(cli.withTXFlags(broadcastCmd...)...)
	RequireTxFailure(t, rsp)

	// test sign batch command by sending same transaction twice

	// TODO: fix batchScanner, not recognizing multiple messages when tested from here

	// // run bank tx send with --generate-only flag
	// bankSendGenCmd = []string{
	// 	"tx", "bank", "send", val1Addr, val2Addr,
	// 	fmt.Sprintf("%d%s", 100, authTestDenom),
	// 	fmt.Sprintf("--fees=%d%s", feeAmount, authTestDenom),
	// 	"--generate-only",
	// }

	// initialOpFile := opFile
	// _ = cli.WithRunErrorMatcher(assertGenOnlyOutput).Run(bankSendGenCmd...)

	// signBatchCmd := []string{"tx", "sign-batch", initialOpFile.Name(), opFile.Name(), "--from=" + val1Addr, "--append"}
	// rsp = cli.RunCommandWithArgs(cli.withTXFlags(signBatchCmd...)...)
}

func TestAuthQueryTxCmds(t *testing.T) {
	// scenario: test query tx and txs commands
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	// get validator addresses
	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	sut.StartChain(t)

	// do a bank transfer and use it for query txs
	feeAmount := "2stake"
	rsp := cli.RunAndWait("tx", "bank", "send", val1Addr, val2Addr, "10000stake", "--fees="+feeAmount)
	RequireTxSuccess(t, rsp)

	// parse values from above tx
	height := gjson.Get(rsp, "height").String()
	txHash := gjson.Get(rsp, "txhash").String()
	accSeq := fmt.Sprintf("%s/%d", val1Addr, gjson.Get(rsp, "tx.auth_info.signer_infos.0.sequence").Int())
	signature := gjson.Get(rsp, "tx.signatures.0").String()

	// test query tx command
	testCases := []struct {
		name string
		args []string
	}{
		{
			"valid query with txhash",
			[]string{txHash},
		},
		{
			"valid query with addr+seq filter",
			[]string{"--type=acc_seq", accSeq},
		},
		{
			"valid query with signature filter",
			[]string{"--type=signature", signature},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := []string{"q", "tx"}
			rsp = cli.CustomQuery(append(cmd, tc.args...)...)
			txHeight := gjson.Get(rsp, "height").String()
			require.Equal(t, height, txHeight)
		})
	}

	// test query txs command
	txsTestCases := []struct {
		name   string
		query  string
		expLen int
	}{
		{
			"fee event happy case",
			fmt.Sprintf("--query=tx.fee='%s'", feeAmount),
			1,
		},
		{
			"no matching fee event",
			"--query=tx.fee='120stake'",
			0,
		},
		{
			"query with tx height",
			fmt.Sprintf("--query=tx.height='%s'", height),
			1,
		},
	}

	for _, tc := range txsTestCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := []string{"q", "txs", tc.query}
			rsp = cli.CustomQuery(cmd...)
			txs := gjson.Get(rsp, "txs").Array()
			require.Equal(t, tc.expLen, len(txs))
		})
	}
}
