//go:build system_test

package systemtests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	systest "cosmossdk.io/systemtests"
)

const (
	authTestDenom = "stake"
)

func TestAuthSignAndBroadcastTxCmd(t *testing.T) {
	// scenario: test auth sign and broadcast commands
	// given a running chain

	systest.Sut.ResetChain(t)
	require.GreaterOrEqual(t, systest.Sut.NodesCount(), 2)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	systest.Sut.StartChain(t)

	var transferAmount, feeAmount int64 = 1000, 1

	// test sign tx command

	// run bank tx send with --generate-only flag
	sendTx := generateBankSendTx(t, cli, val1Addr, val2Addr, transferAmount, feeAmount, "")
	txFile := systest.StoreTempFile(t, []byte(fmt.Sprintf("%s\n", sendTx)))

	// query node0 account details
	signTxCmd := []string{"tx", "sign", txFile.Name(), "--from=" + val1Addr, "--chain-id=" + cli.ChainID()}
	testSignTxBroadcast(t, cli, signTxCmd, "sign tx", val1Addr, val2Addr, transferAmount, feeAmount)

	// test broadcast with empty public key in signed tx
	rsp := cli.RunCommandWithArgs(cli.WithTXFlags("tx", "sign", txFile.Name(), "--from="+val1Addr)...)
	updated, err := sjson.Set(rsp, "auth_info.signer_infos.0.public_key", nil)
	require.NoError(t, err)
	newSignFile := systest.StoreTempFile(t, []byte(updated))

	broadcastCmd := []string{"tx", "broadcast", newSignFile.Name()}
	rsp = cli.WithRunErrorsIgnored().RunCommandWithArgs(cli.WithTXFlags(broadcastCmd...)...) //  // ignore run errors, as comet exit with 1 on rpc errors
	systest.RequireTxFailure(t, rsp)

	// test sign-batch tx command

	// generate another bank send tx with less amount
	newAmount := int64(100)
	sendTx2 := generateBankSendTx(t, cli, val1Addr, val2Addr, newAmount, feeAmount, "")
	tx2File := systest.StoreTempFile(t, []byte(fmt.Sprintf("%s\n", sendTx2)))

	signBatchCmd := []string{"tx", "sign-batch", txFile.Name(), tx2File.Name(), "--from=" + val1Addr, "--chain-id=" + cli.ChainID()}
	sendAmount := transferAmount + newAmount
	fees := feeAmount * 2

	testSignTxBroadcast(t, cli, signBatchCmd, "sign-batch tx", val1Addr, val2Addr, sendAmount, fees)
}

func testSignTxBroadcast(t *testing.T, cli *systest.CLIWrapper, txCmd []string, prefix, fromAddr, toAddr string, amount, fees int64) {
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
			rsp = cli.RunCommandWithArgs(cli.WithKeyringFlags(cmd...)...)

			signatures := gjson.Get(rsp, "signatures").Array()
			require.Len(t, signatures, 1)

			signFile := systest.StoreTempFile(t, []byte(rsp))

			// validate signature
			rsp = cli.RunCommandWithArgs(cli.WithKeyringFlags("tx", "validate-signatures", signFile.Name(), "--from="+fromAddr, "--chain-id="+cli.ChainID())...)
			require.Contains(t, rsp, "[OK]")

			// run broadcast tx command
			broadcastCmd := []string{"tx", "broadcast", signFile.Name()}
			rsp = cli.RunAndWait(broadcastCmd...)
			systest.RequireTxSuccess(t, rsp)

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

	systest.Sut.ResetChain(t)
	require.GreaterOrEqual(t, systest.Sut.NodesCount(), 2)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	systest.Sut.StartChain(t)

	// do a bank transfer and use it for query txs
	feeAmount := "2stake"
	rsp := cli.RunAndWait("tx", "bank", "send", val1Addr, val2Addr, "10000stake", "--fees="+feeAmount)
	systest.RequireTxSuccess(t, rsp)

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

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

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

	out := cli.RunCommandWithArgs(cli.WithKeyringFlags("keys", "add", "multi", "--multisig=acc1,acc2,acc3", "--multisig-threshold=2")...)
	multiAddr := gjson.Get(out, "address").String()
	require.NotEmpty(t, multiAddr)

	systest.Sut.StartChain(t)

	// fund multisig address some amount
	var initialAmount, transferAmount, feeAmount int64 = 10000, 100, 1
	_ = cli.FundAddress(multiAddr, fmt.Sprintf("%d%s", initialAmount, authTestDenom))

	multiAddrBal := cli.QueryBalance(multiAddr, authTestDenom)
	require.Equal(t, initialAmount, multiAddrBal)

	// test multisign tx command

	// run bank tx send with --generate-only flag
	sendTx := generateBankSendTx(t, cli, multiAddr, valAddr, transferAmount, feeAmount, "")
	txFile := systest.StoreTempFile(t, []byte(sendTx))

	signTxCmd := cli.WithKeyringFlags("tx", "sign", txFile.Name(), "--multisig="+multiAddr, "--chain-id="+cli.ChainID())
	multiSignTxCmd := cli.WithKeyringFlags("tx", "multisign", txFile.Name(), "multi", "--chain-id="+cli.ChainID())

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
	multiTxFile := systest.StoreTempFile(t, []byte(multiSendTx))

	signBatchTxCmd := cli.WithKeyringFlags("tx", "sign-batch", multiTxFile.Name(), "--multisig="+multiAddr, "--signature-only", "--chain-id="+cli.ChainID())
	multiSignBatchTxCmd := cli.WithKeyringFlags("tx", "multisign-batch", multiTxFile.Name(), "multi", "--chain-id="+cli.ChainID())

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

func generateBankSendTx(t *testing.T, cli *systest.CLIWrapper, fromAddr, toAddr string, amount, fees int64, memo string) string {
	t.Helper()

	bankSendGenCmd := []string{
		"tx", "bank", "send", fromAddr, toAddr,
		fmt.Sprintf("%d%s", amount, authTestDenom),
		fmt.Sprintf("--fees=%d%s", fees, authTestDenom),
		"--generate-only",
		"--note=" + memo,
	}

	return cli.RunCommandWithArgs(cli.WithTXFlags(bankSendGenCmd...)...)
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

func testMultisigTxBroadcast(t *testing.T, cli *systest.CLIWrapper, i multiSigTxInput) {
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
				signFile := systest.StoreTempFile(t, []byte(rsp))
				cmd = append(cmd, signFile.Name())
			}
			rsp := cli.RunCommandWithArgs(cmd...)
			multiSignFile := systest.StoreTempFile(t, []byte(rsp))

			// run broadcast tx command
			broadcastCmd := []string{"tx", "broadcast", multiSignFile.Name()}
			if tc.expErrMsg != "" {
				rsp = cli.WithRunErrorsIgnored().RunCommandWithArgs(cli.WithTXFlags(broadcastCmd...)...) //  // ignore run errors, as comet exit with 1 on rpc errors
				systest.RequireTxFailure(t, rsp)
				require.Contains(t, rsp, tc.expErrMsg)
				return
			}

			rsp = cli.RunAndWait(broadcastCmd...)
			systest.RequireTxSuccess(t, rsp)

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

	systest.Sut.ResetChain(t)
	require.GreaterOrEqual(t, systest.Sut.NodesCount(), 2)

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator addresses
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	val2Addr := cli.GetKeyAddr("node1")
	require.NotEmpty(t, val2Addr)

	systest.Sut.StartChain(t)

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

			_ = cli.WithRunErrorMatcher(assertTxOutput).Run(cli.WithTXFlags(cmd...)...)
		})
	}
}

func TestTxEncodeandDecodeAndQueries(t *testing.T) {
	// scenario: test tx encode and decode commands

	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)

	// get validator address
	val1Addr := cli.GetKeyAddr("node0")
	require.NotEmpty(t, val1Addr)

	memoText := "testmemo"
	sendTx := generateBankSendTx(t, cli, val1Addr, val1Addr, 100, 1, memoText)
	txFile := systest.StoreTempFile(t, []byte(sendTx))

	// run encode command
	encodedText := cli.RunCommandWithArgs("tx", "encode", txFile.Name())

	// check transaction decodes as expected
	decodedTx := cli.RunCommandWithArgs("tx", "decode", encodedText)
	require.Equal(t, gjson.Get(decodedTx, "body.memo").String(), memoText)

	systest.Sut.StartChain(t)
	// Test Gateway queries
	addr := "cosmos1q6cc9u0x5r3fkjcex0rgxee5qlu86w8rh2ypaj"
	addrBytesURLEncoded := "BrGC8eag4ptLGTPGg2c0B%2Fh9OOM%3D"
	addrBytes := "BrGC8eag4ptLGTPGg2c0B/h9OOM="

	baseurl := systest.Sut.APIAddress()
	stringToBytesPath := baseurl + "/cosmos/auth/v1beta1/bech32/encode/%s"
	bytesToStringPath := baseurl + "/cosmos/auth/v1beta1/bech32/%s"
	bytesToStringPath2 := baseurl + "/cosmos/auth/v1beta1/bech32/decode/%s"
	testCases := []systest.RestTestCase{
		{
			Name:    "convert string to bytes",
			Url:     fmt.Sprintf(stringToBytesPath, addr),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"address_bytes":"%s"}`, addrBytes),
		},
		{
			Name:    "convert bytes to string",
			Url:     fmt.Sprintf(bytesToStringPath, addrBytesURLEncoded),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"address_string":"%s"}`, addr),
		},
		{
			Name:    "convert bytes to string other endpoint",
			Url:     fmt.Sprintf(bytesToStringPath2, addrBytesURLEncoded),
			ExpCode: http.StatusOK,
			ExpOut:  fmt.Sprintf(`{"address_string":"%s"}`, addr),
		},
		{
			Name:    "should fail with bad address",
			Url:     fmt.Sprintf(stringToBytesPath, "aslkdjglksdfhjlksdjfhlkjsdfh"),
			ExpCode: http.StatusInternalServerError,
			ExpOut:  `{"code":2,"message":"decoding bech32 failed: invalid separator index -1","details":[]}`,
		},
		{
			Name:    "should fail with bad bytes",
			Url:     fmt.Sprintf(bytesToStringPath, "f"),
			ExpCode: http.StatusBadRequest,
			ExpOut:  `{"code":3,"message":"failed to populate field address_bytes with value f: illegal base64 data at input byte 0","details":[]}`,
		},
	}
	systest.RunRestQueries(t, testCases...)
}

func TestTxWithFeePayer(t *testing.T) {
	// Scenario:
	// send a tx with FeePayer without his signature
	// check tx fails
	// send tx with feePayers signature
	// check tx executed ok
	// check fees had been deducted from feePayers balance

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose).WithRunErrorsIgnored()

	// add sender and feePayer accounts
	senderAddr := cli.AddKey("sender")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000stake"},
	)
	feePayerAddr := cli.AddKey("feePayer")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", feePayerAddr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	// send a tx with FeePayer without his signature
	rsp := cli.WithRunErrorsIgnored().RunCommandWithArgs(cli.WithTXFlags(
		"tx", "bank", "send", senderAddr, "cosmos108jsm625z3ejy63uef2ke7t67h6nukt4ty93nr", "1000stake", "--fees", "1000000stake", "--fee-payer", feePayerAddr,
	)...) // ignore run errors, as comet exit with 1 on rpc errors
	systest.RequireTxFailure(t, rsp, "invalid number of signatures")

	// send tx with feePayers signature
	rsp = cli.RunCommandWithArgs(cli.WithTXFlags(
		"tx", "bank", "send", senderAddr, "cosmos108jsm625z3ejy63uef2ke7t67h6nukt4ty93nr", "1000stake", "--fees", "1000000stake", "--fee-payer", feePayerAddr, "--generate-only",
	)...)
	tempFile := systest.StoreTempFile(t, []byte(rsp))

	rsp = cli.RunCommandWithArgs(cli.WithTXFlags(
		"tx", "sign", tempFile.Name(), "--from", senderAddr, "--sign-mode", "amino-json",
	)...)
	tempFile = systest.StoreTempFile(t, []byte(rsp))

	rsp = cli.RunCommandWithArgs(cli.WithTXFlags(
		"tx", "sign", tempFile.Name(), "--from", feePayerAddr, "--sign-mode", "amino-json",
	)...)
	tempFile = systest.StoreTempFile(t, []byte(rsp))

	rsp = cli.RunAndWait([]string{"tx", "broadcast", tempFile.Name()}...)
	systest.RequireTxSuccess(t, rsp)

	// Query to check fee has been deducted from feePayer
	balance := cli.QueryBalance(feePayerAddr, authTestDenom)
	assert.Equal(t, balance, int64(9000000))
}
