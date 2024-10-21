//go:build system_test

package systemtests

import (
	"fmt"
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
	require.GreaterOrEqual(t, sut.NodesCount(), 2)

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	sut.StartChain(t)

	var transferAmount, feeAmount int64 = 1000, 1

	// test sign tx command

	// run bank tx send with --generate-only flag
	sendTx := generateBankSendTx(t, cli, val1Addr, val2Addr, transferAmount, feeAmount, "")
	txFile := StoreTempFile(t, []byte(fmt.Sprintf("%s\n", sendTx)))

	// query node0 account details
	signTxCmd := []string{"tx", "sign", txFile.Name(), "--from=" + val1Addr, "--chain-id=" + cli.chainID}
	testSignTxBroadcast(t, cli, signTxCmd, "sign tx", val1Addr, val2Addr, transferAmount, feeAmount)

	// test broadcast with empty public key in signed tx
	rsp := cli.RunCommandWithArgs(cli.withTXFlags("tx", "sign", txFile.Name(), "--from="+val1Addr)...)
	updated, err := sjson.Set(rsp, "auth_info.signer_infos.0.public_key", nil)
	require.NoError(t, err)
	newSignFile := StoreTempFile(t, []byte(updated))

	broadcastCmd := []string{"tx", "broadcast", newSignFile.Name()}
	rsp = cli.RunCommandWithArgs(cli.withTXFlags(broadcastCmd...)...)
	RequireTxFailure(t, rsp)

	// test sign-batch tx command

	// generate another bank send tx with less amount
	newAmount := int64(100)
	sendTx2 := generateBankSendTx(t, cli, val1Addr, val2Addr, newAmount, feeAmount, "")
	tx2File := StoreTempFile(t, []byte(fmt.Sprintf("%s\n", sendTx2)))

	signBatchCmd := []string{"tx", "sign-batch", txFile.Name(), tx2File.Name(), "--from=" + val1Addr, "--chain-id=" + cli.chainID}
	sendAmount := transferAmount + newAmount
	fees := feeAmount * 2

	testSignTxBroadcast(t, cli, signBatchCmd, "sign-batch tx", val1Addr, val2Addr, sendAmount, fees)
}

func testSignTxBroadcast(t *testing.T, cli *CLIWrapper, txCmd []string, prefix, fromAddr, toAddr string, amount, fees int64) {
	t.Helper()

	fromAddrBal := cli.QueryBalance(fromAddr, authTestDenom)
	toAddrBal := cli.QueryBalance(toAddr, authTestDenom)

	// query from account details
	rsp := cli.CustomQuery("q", "auth", "accounts")
	details := gjson.Get(rsp, fmt.Sprintf("accounts.#(value.address==%s).value", fromAddr)).String()
	accSeq := gjson.Get(details, "sequence").Int()
	accNum := gjson.Get(details, "account_number").Int()

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

	for _, tc := range testCases {
		t.Run(prefix+"-"+tc.name, func(t *testing.T) {
			cmd := append(txCmd, tc.extraArgs...)

			// run tx sign command and verify signatures count
			rsp = cli.RunCommandWithArgs(cli.withKeyringFlags(cmd...)...)

			signatures := gjson.Get(rsp, "signatures").Array()
			require.Len(t, signatures, 1)

			signFile := StoreTempFile(t, []byte(rsp))

			// validate signature
			rsp = cli.RunCommandWithArgs(cli.withKeyringFlags("tx", "validate-signatures", signFile.Name(), "--from="+fromAddr, "--chain-id="+cli.chainID)...)
			require.Contains(t, rsp, "[OK]")

			// run broadcast tx command
			broadcastCmd := []string{"tx", "broadcast", signFile.Name()}
			rsp = cli.RunAndWait(broadcastCmd...)
			RequireTxSuccess(t, rsp)

			// query balance and confirm transaction
			expVal1Bal := fromAddrBal - amount - fees
			fromAddrBal = cli.QueryBalance(fromAddr, authTestDenom)
			require.Equal(t, expVal1Bal, fromAddrBal)

			expVal2Bal := toAddrBal + amount
			toAddrBal = cli.QueryBalance(toAddr, authTestDenom)
			require.Equal(t, expVal2Bal, toAddrBal)
		})
	}
}

func TestAuthQueryTxCmds(t *testing.T) {
	// scenario: test query tx and txs commands
	// given a running chain

	sut.ResetChain(t)
	require.GreaterOrEqual(t, sut.NodesCount(), 2)

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

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

func TestAuthMultisigTxCmds(t *testing.T) {
	// scenario: test auth multisig related tx commands
	// given a running chain

	sut.ResetChain(t)
	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	valAddr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// add new keys for multisig
	acc1Addr := cli.AddKey("acc1")
	require.NotEmpty(t, acc1Addr)

	acc2Addr := cli.AddKey("acc2")
	require.NotEqual(t, acc1Addr, acc2Addr)

	acc3Addr := cli.AddKey("acc3")
	require.NotEqual(t, acc1Addr, acc3Addr)

	out := cli.RunCommandWithArgs(cli.withKeyringFlags("keys", "add", "multi", "--multisig=acc1,acc2,acc3", "--multisig-threshold=2")...)
	multiAddr := gjson.Get(out, "address").String()
	require.NotEmpty(t, multiAddr)

	sut.StartChain(t)

	// fund multisig address some amount
	var initialAmount, transferAmount, feeAmount int64 = 10000, 100, 1
	_ = cli.FundAddress(multiAddr, fmt.Sprintf("%d%s", initialAmount, authTestDenom))

	multiAddrBal := cli.QueryBalance(multiAddr, authTestDenom)
	require.Equal(t, initialAmount, multiAddrBal)

	// test multisign tx command

	// run bank tx send with --generate-only flag
	sendTx := generateBankSendTx(t, cli, multiAddr, valAddr, transferAmount, feeAmount, "")
	txFile := StoreTempFile(t, []byte(sendTx))

	signTxCmd := cli.withKeyringFlags("tx", "sign", txFile.Name(), "--multisig="+multiAddr, "--chain-id="+cli.chainID)
	multiSignTxCmd := cli.withKeyringFlags("tx", "multisign", txFile.Name(), "multi", "--chain-id="+cli.chainID)

	testMultisigTxBroadcast(t, cli, multiSigTxInput{
		"multisign",
		multiAddr,
		valAddr,
		acc1Addr,
		acc2Addr,
		acc3Addr,
		signTxCmd,
		multiSignTxCmd,
		transferAmount,
		feeAmount,
	})

	// test multisign-batch tx command

	// generate two send transactions in single file
	multiSendTx := fmt.Sprintf("%s\n%s", sendTx, sendTx)
	multiTxFile := StoreTempFile(t, []byte(multiSendTx))

	signBatchTxCmd := cli.withKeyringFlags("tx", "sign-batch", multiTxFile.Name(), "--multisig="+multiAddr, "--signature-only", "--chain-id="+cli.chainID)
	multiSignBatchTxCmd := cli.withKeyringFlags("tx", "multisign-batch", multiTxFile.Name(), "multi", "--chain-id="+cli.chainID)

	// as we done couple of bank transactions as batch,
	// transferred amount will be twice
	sendAmount := transferAmount * 2
	fees := feeAmount * 2

	testMultisigTxBroadcast(t, cli, multiSigTxInput{
		"multisign-batch",
		multiAddr,
		valAddr,
		acc1Addr,
		acc2Addr,
		acc3Addr,
		signBatchTxCmd,
		multiSignBatchTxCmd,
		sendAmount,
		fees,
	})
}

func generateBankSendTx(t *testing.T, cli *CLIWrapper, fromAddr, toAddr string, amount, fees int64, memo string) string {
	t.Helper()

	bankSendGenCmd := []string{
		"tx", "bank", "send", fromAddr, toAddr,
		fmt.Sprintf("%d%s", amount, authTestDenom),
		fmt.Sprintf("--fees=%d%s", fees, authTestDenom),
		"--generate-only",
		"--note=" + memo,
	}

	return cli.RunCommandWithArgs(cli.withTXFlags(bankSendGenCmd...)...)
}

type multiSigTxInput struct {
	prefix         string
	multiAddr      string
	toAddr         string
	acc1Addr       string
	acc2Addr       string
	acc3Addr       string
	signCmd        []string
	multiSignCmd   []string
	transferAmount int64
	feeAmount      int64
}

func testMultisigTxBroadcast(t *testing.T, cli *CLIWrapper, i multiSigTxInput) {
	t.Helper()

	multiAddrBal := cli.QueryBalance(i.multiAddr, authTestDenom)
	toAddrBal := cli.QueryBalance(i.toAddr, authTestDenom)

	testCases := []struct {
		name        string
		signingAccs []string
		expErrMsg   string
	}{
		{
			"minimum threshold not reached",
			[]string{i.acc1Addr},
			"signature size is incorrect",
		},
		{
			"valid tx with two signers",
			[]string{i.acc1Addr, i.acc2Addr},
			"",
		},
		{
			"valid tx with three signed files",
			[]string{i.acc1Addr, i.acc2Addr, i.acc3Addr},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(i.prefix+"-"+tc.name, func(t *testing.T) {
			// append signed files to multisign command
			cmd := i.multiSignCmd
			for _, acc := range tc.signingAccs {
				rsp := cli.RunCommandWithArgs(append(i.signCmd, "--from="+acc)...)
				signFile := StoreTempFile(t, []byte(rsp))
				cmd = append(cmd, signFile.Name())
			}
			rsp := cli.RunCommandWithArgs(cmd...)
			multiSignFile := StoreTempFile(t, []byte(rsp))

			// run broadcast tx command
			broadcastCmd := []string{"tx", "broadcast", multiSignFile.Name()}
			if tc.expErrMsg != "" {
				rsp = cli.RunCommandWithArgs(cli.withTXFlags(broadcastCmd...)...)
				RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
				return
			}

			rsp = cli.RunAndWait(broadcastCmd...)
			RequireTxSuccess(t, rsp)

			// query balance and confirm transaction
			expMultiBal := multiAddrBal - i.transferAmount - i.feeAmount
			multiAddrBal = cli.QueryBalance(i.multiAddr, authTestDenom)
			require.Equal(t, expMultiBal, multiAddrBal)

			expVal2Bal := toAddrBal + i.transferAmount
			toAddrBal = cli.QueryBalance(i.toAddr, authTestDenom)
			require.Equal(t, expVal2Bal, toAddrBal)
		})
	}
}

func TestAuxSigner(t *testing.T) {
	// scenario: test tx with direct aux sign mode
	// given a running chain

	sut.ResetChain(t)
	require.GreaterOrEqual(t, sut.NodesCount(), 2)

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	sut.StartChain(t)

	bankSendCmd := []string{"tx", "bank", "send", val1Addr, val2Addr, "10000stake", "--from=" + val1Addr}

	testCases := []struct {
		name      string
		args      []string
		expErrMsg string
	}{
		{
			"error with SIGN_MODE_DIRECT_AUX and --aux unset",
			[]string{
				"--sign-mode=direct-aux",
			},
			"cannot sign with SIGN_MODE_DIRECT_AUX",
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only set (ignores generate-only)",
			[]string{
				"--sign-mode=direct-aux",
				"--generate-only",
			},
			"",
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only, tip flag set",
			[]string{
				"--sign-mode=direct-aux",
				"--generate-only",
				"--tip=10stake",
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := append(bankSendCmd, tc.args...)
			assertTxOutput := func(_ assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
				require.Len(t, gotOutputs, 1)
				output := gotOutputs[0].(string)
				if tc.expErrMsg != "" {
					require.Contains(t, output, tc.expErrMsg)
				} else {
					require.Len(t, gjson.Get(output, "body.messages").Array(), 1)
				}
				return false
			}

			_ = cli.WithRunErrorMatcher(assertTxOutput).Run(cli.withTXFlags(cmd...)...)
		})
	}
}

func TestTxEncodeandDecode(t *testing.T) {
	// scenario: test tx encode and decode commands

	cli := NewCLIWrapper(t, sut, verbose)

	// get validator address
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	memoText := "testmemo"
	sendTx := generateBankSendTx(t, cli, val1Addr, val1Addr, 100, 1, memoText)
	txFile := StoreTempFile(t, []byte(sendTx))

	// run encode command
	encodedText := cli.RunCommandWithArgs("tx", "encode", txFile.Name())

	// check transaction decodes as expected
	decodedTx := cli.RunCommandWithArgs("tx", "decode", encodedText)
	require.Equal(t, gjson.Get(decodedTx, "body.memo").String(), memoText)
}
