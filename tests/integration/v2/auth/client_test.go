package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	govtestutil "cosmossdk.io/x/gov/client/testutil"
	govtypes "cosmossdk.io/x/gov/types/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtestutil "github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestCLIValidateSignatures(t *testing.T) {
	f := createTestSuite(t)

	sendTokens := sdk.NewCoins(
		sdk.NewCoin("testtoken", math.NewInt(10)),
		sdk.NewCoin("stake", math.NewInt(10)))

	res, err := createBankMsg(f.clientCtx, f.val, f.val, sendTokens, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	// write  unsigned tx to file
	unsignedTx := testutil.WriteToNewTempFile(t, res.String())
	defer unsignedTx.Close()

	res, err = authtestutil.TxSignExec(f.clientCtx, f.val, unsignedTx.Name())
	require.NoError(t, err)
	signedTx, err := f.clientCtx.TxConfig.TxJSONDecoder()(res.Bytes())
	require.NoError(t, err)

	signedTxFile := testutil.WriteToNewTempFile(t, res.String())
	defer signedTxFile.Close()
	txBuilder, err := f.clientCtx.TxConfig.WrapTxBuilder(signedTx)
	require.NoError(t, err)
	_, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)

	txBuilder.SetMemo("MODIFIED TX")
	bz, err := f.clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	modifiedTxFile := testutil.WriteToNewTempFile(t, string(bz))
	defer modifiedTxFile.Close()

	_, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, modifiedTxFile.Name())
	require.EqualError(t, err, "signatures validation failed")
}

func TestCLISignBatch(t *testing.T) {
	f := createTestSuite(t)

	sendTokens := sdk.NewCoins(
		sdk.NewCoin("testtoken", math.NewInt(10)),
		sdk.NewCoin("stake", math.NewInt(10)),
	)

	generatedStd, err := createBankMsg(f.clientCtx, f.val, f.val, sendTokens, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	outputFile := testutil.WriteToNewTempFile(t, strings.Repeat(generatedStd.String(), 3))
	defer outputFile.Close()
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)

	// sign-batch file - offline is set but account-number and sequence are not
	_, err = authtestutil.TxSignBatchExec(f.clientCtx, f.val, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, f.clientCtx.ChainID), "--offline")
	require.EqualError(t, err, "required flag(s) \"account-number\", \"sequence\" not set")

	// sign-batch file - offline and sequence is set but account-number is not set
	_, err = authtestutil.TxSignBatchExec(f.clientCtx, f.val, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, f.clientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagSequence, "1"), "--offline")
	require.EqualError(t, err, "required flag(s) \"account-number\" not set")

	// sign-batch file - offline and account-number is set but sequence is not set
	_, err = authtestutil.TxSignBatchExec(f.clientCtx, f.val, outputFile.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, f.clientCtx.ChainID), fmt.Sprintf("--%s=%s", flags.FlagAccountNumber, "1"), "--offline")
	require.EqualError(t, err, "required flag(s) \"sequence\" not set")
}

func TestCLISignBatchTotalFees(t *testing.T) {
	f := createTestSuite(t)

	txCfg := f.clientCtx.TxConfig
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)

	testCases := []struct {
		name            string
		args            []string
		numTransactions int
		denom           string
	}{
		{
			"Offline batch-sign one transaction",
			[]string{"--offline", "--account-number", "1", "--sequence", "1", "--append"},
			1,
			"stake",
		},
		{
			"Offline batch-sign two transactions",
			[]string{"--offline", "--account-number", "1", "--sequence", "1", "--append"},
			2,
			"stake",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create multiple transactions and write them to separate files
			sendTokens := sdk.NewCoin(tc.denom, sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))
			expectedBatchedTotalFee := int64(0)
			txFiles := make([]string, tc.numTransactions)
			for i := 0; i < tc.numTransactions; i++ {
				tx, err := createBankMsg(f.clientCtx, f.val, f.val, sdk.NewCoins(sendTokens), clitestutil.TestTxConfig{GenOnly: true})
				require.NoError(t, err)
				txFile := testutil.WriteToNewTempFile(t, tx.String()+"\n")
				txFiles[i] = txFile.Name()

				unsignedTx, err := txCfg.TxJSONDecoder()(tx.Bytes())
				require.NoError(t, err)
				txBuilder, err := txCfg.WrapTxBuilder(unsignedTx)
				require.NoError(t, err)
				expectedBatchedTotalFee += txBuilder.GetTx().GetFee().AmountOf(tc.denom).Int64()
				err = txFile.Close()
				require.NoError(t, err)
			}

			// Test batch sign
			batchSignArgs := append([]string{fmt.Sprintf("--from=%s", f.val.String())}, append(txFiles, tc.args...)...)
			signedTx, err := clitestutil.ExecTestCLICmd(f.clientCtx, authcli.GetSignBatchCommand(), batchSignArgs)
			require.NoError(t, err)
			signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
			require.NoError(t, err)
			txBuilder, err := txCfg.WrapTxBuilder(signedFinalTx)
			require.NoError(t, err)
			finalTotalFee := txBuilder.GetTx().GetFee()
			require.Equal(t, expectedBatchedTotalFee, finalTotalFee.AmountOf(tc.denom).Int64())
		})
	}
}

func TestCLISendGenerateSignAndBroadcast(t *testing.T) {
	f := createTestSuite(t)

	sendTokens := sdk.NewCoin("stake", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	normalGeneratedTx, err := createBankMsg(f.clientCtx, f.val, f.val, sdk.NewCoins(sendTokens), clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	txCfg := f.clientCtx.TxConfig

	normalGeneratedStdTx, err := txCfg.TxJSONDecoder()(normalGeneratedTx.Bytes())
	require.NoError(t, err)
	txBuilder, err := txCfg.WrapTxBuilder(normalGeneratedStdTx)
	require.NoError(t, err)
	require.Equal(t, txBuilder.GetTx().GetGas(), uint64(flags.DefaultGasLimit))
	require.Equal(t, len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err := txBuilder.GetTx().GetSignaturesV2()
	require.NoError(t, err)
	require.Equal(t, 0, len(sigs))

	// Test generate sendTx with --gas=$amount
	limitedGasGeneratedTx, err := createBankMsg(f.clientCtx, f.val, f.val,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{GenOnly: true, Gas: 100},
	)
	require.NoError(t, err)

	limitedGasStdTx, err := txCfg.TxJSONDecoder()(limitedGasGeneratedTx.Bytes())
	require.NoError(t, err)
	txBuilder, err = txCfg.WrapTxBuilder(limitedGasStdTx)
	require.NoError(t, err)
	require.Equal(t, txBuilder.GetTx().GetGas(), uint64(100))
	require.Equal(t, len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	require.NoError(t, err)
	require.Equal(t, 0, len(sigs))

	// Test generate sendTx, estimate gas
	finalGeneratedTx, err := createBankMsg(f.clientCtx, f.val, f.val,
		sdk.NewCoins(sendTokens), clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	finalStdTx, err := txCfg.TxJSONDecoder()(finalGeneratedTx.Bytes())
	require.NoError(t, err)
	txBuilder, err = txCfg.WrapTxBuilder(finalStdTx)
	require.NoError(t, err)
	require.Equal(t, uint64(flags.DefaultGasLimit), txBuilder.GetTx().GetGas())
	require.Equal(t, len(finalStdTx.GetMsgs()), 1)

	// Write the output to disk
	unsignedTxFile := testutil.WriteToNewTempFile(t, finalGeneratedTx.String())
	defer unsignedTxFile.Close()

	// Test validate-signatures
	res, err := authtestutil.TxValidateSignaturesExec(f.clientCtx, unsignedTxFile.Name())
	require.EqualError(t, err, "signatures validation failed")
	require.Contains(t, res.String(), fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", f.val.String()))

	// Test sign

	// Does not work in offline mode
	_, err = authtestutil.TxSignExec(f.clientCtx, f.val, unsignedTxFile.Name(), "--offline")
	require.EqualError(t, err, "required flag(s) \"account-number\", \"sequence\" not set")

	// But works offline if we set account number and sequence
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)
	_, err = authtestutil.TxSignExec(f.clientCtx, f.val, unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	require.NoError(t, err)

	// Sign transaction
	signedTx, err := authtestutil.TxSignExec(f.clientCtx, f.val, unsignedTxFile.Name())
	require.NoError(t, err)
	signedFinalTx, err := txCfg.TxJSONDecoder()(signedTx.Bytes())
	require.NoError(t, err)
	txBuilder, err = f.clientCtx.TxConfig.WrapTxBuilder(signedFinalTx)
	require.NoError(t, err)
	require.Equal(t, len(txBuilder.GetTx().GetMsgs()), 1)
	sigs, err = txBuilder.GetTx().GetSignaturesV2()
	require.NoError(t, err)
	require.Equal(t, 1, len(sigs))
	signers, err := txBuilder.GetTx().GetSigners()
	require.NoError(t, err)
	require.Equal(t, []byte(f.val), signers[0])

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(t, signedTx.String())
	defer signedTxFile.Close()

	// validate Signature
	res, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)
	require.True(t, strings.Contains(res.String(), "[OK]"))

	// Test broadcast

	// Does not work in offline mode
	_, err = authtestutil.TxBroadcastExec(f.clientCtx, signedTxFile.Name(), "--offline")
	require.EqualError(t, err, "cannot broadcast tx during offline mode")

	// Broadcast correct transaction.
	f.clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authtestutil.TxBroadcastExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)
}

func TestCLIMultisignInsufficientCosigners(t *testing.T) {
	f := createTestSuite(t)

	// Fetch account and a multisig info
	account1, err := f.clientCtx.Keyring.Key("newAccount1")
	require.NoError(t, err)

	multisigRecord, err := f.clientCtx.Keyring.Key("multi")
	require.NoError(t, err)

	addr, err := multisigRecord.GetAddress()
	require.NoError(t, err)
	// Send coins from validator to multisig.
	_, err = createBankMsg(
		f.clientCtx,
		f.val,
		addr,
		sdk.NewCoins(
			sdk.NewInt64Coin("stake", 10),
		),
		clitestutil.TestTxConfig{},
	)
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   f.val.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 5)),
	}

	multiGeneratedTx, err := clitestutil.SubmitTestTx(
		f.clientCtx,
		msgSend,
		addr,
		clitestutil.TestTxConfig{
			GenOnly: true,
		})
	require.NoError(t, err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(t, multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Multisign, sign with one signature
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)
	addr1, err := account1.GetAddress()
	require.NoError(t, err)
	account1Signature, err := authtestutil.TxSignExec(f.clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.NoError(t, err)

	sign1File := testutil.WriteToNewTempFile(t, account1Signature.String())
	defer sign1File.Close()

	multiSigWith1Signature, err := authtestutil.TxMultiSignExec(f.clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name())
	require.NoError(t, err)

	// Save tx to file
	multiSigWith1SignatureFile := testutil.WriteToNewTempFile(t, multiSigWith1Signature.String())
	defer multiSigWith1SignatureFile.Close()

	_, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, multiSigWith1SignatureFile.Name())
	require.Error(t, err)
}

func TestCLIEncode(t *testing.T) {
	f := createTestSuite(t)

	sendTokens := sdk.NewCoin("stake", sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction))

	normalGeneratedTx, err := createBankMsg(
		f.clientCtx, f.val, f.val,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{
			GenOnly: true,
			Memo:    "deadbeef",
		},
	)
	require.NoError(t, err)
	savedTxFile := testutil.WriteToNewTempFile(t, normalGeneratedTx.String())
	defer savedTxFile.Close()

	// Encode
	encodeExec, err := authtestutil.TxEncodeExec(f.clientCtx, savedTxFile.Name())
	require.NoError(t, err)
	trimmedBase64 := strings.Trim(encodeExec.String(), "\"\n")

	// Check that the transaction decodes as expected
	decodedTx, err := authtestutil.TxDecodeExec(f.clientCtx, trimmedBase64)
	require.NoError(t, err)

	txCfg := f.clientCtx.TxConfig
	theTx, err := txCfg.TxJSONDecoder()(decodedTx.Bytes())
	require.NoError(t, err)
	txBuilder, err := f.clientCtx.TxConfig.WrapTxBuilder(theTx)
	require.NoError(t, err)
	require.Equal(t, "deadbeef", txBuilder.GetTx().GetMemo())
}

func TestCLIMultisignSortSignatures(t *testing.T) {
	f := createTestSuite(t)

	// Generate 2 accounts and a multisig.
	account1, err := f.clientCtx.Keyring.Key("newAccount1")
	require.NoError(t, err)

	account2, err := f.clientCtx.Keyring.Key("newAccount2")
	require.NoError(t, err)

	multisigRecord, err := f.clientCtx.Keyring.Key("multi")
	require.NoError(t, err)

	// Generate dummy account which is not a part of multisig.
	dummyAcc, err := f.clientCtx.Keyring.Key("dummyAccount")
	require.NoError(t, err)

	addr, err := multisigRecord.GetAddress()
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   f.val.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 5)),
	}

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.SubmitTestTx(f.clientCtx, msgSend, addr, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(t, multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	// Sign with account1
	addr1, err := account1.GetAddress()
	require.NoError(t, err)
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtestutil.TxSignExec(f.clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.NoError(t, err)

	sign1File := testutil.WriteToNewTempFile(t, account1Signature.String())
	defer sign1File.Close()

	// Sign with account2
	addr2, err := account2.GetAddress()
	require.NoError(t, err)
	account2Signature, err := authtestutil.TxSignExec(f.clientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.NoError(t, err)

	sign2File := testutil.WriteToNewTempFile(t, account2Signature.String())
	defer sign2File.Close()

	// Sign with dummy account
	dummyAddr, err := dummyAcc.GetAddress()
	require.NoError(t, err)
	_, err = authtestutil.TxSignExec(f.clientCtx, dummyAddr, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.Error(t, err)
	require.Contains(t, err.Error(), "signing key is not a part of multisig key")

	multiSigWith2Signatures, err := authtestutil.TxMultiSignExec(f.clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	require.NoError(t, err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(t, multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)

	f.clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authtestutil.TxBroadcastExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)
}

func TestSignWithMultisig(t *testing.T) {
	f := createTestSuite(t)

	// Generate a account for signing.
	account1, err := f.clientCtx.Keyring.Key("newAccount1")
	require.NoError(t, err)

	addr1, err := account1.GetAddress()
	require.NoError(t, err)

	// Create an address that is not in the keyring, will be used to simulate `--multisig`
	multisig := "cosmos1hd6fsrvnz6qkp87s3u86ludegq97agxsdkwzyh"
	_, err = f.clientCtx.AddressCodec.StringToBytes(multisig)
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: f.val.String(),
		ToAddress:   f.val.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 5)),
	}

	// Generate a transaction for testing --multisig with an address not in the keyring.
	multisigTx, err := clitestutil.SubmitTestTx(f.clientCtx, msgSend, f.val, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	// Save multi tx to file
	multiGeneratedTx2File := testutil.WriteToNewTempFile(t, multisigTx.String())
	defer multiGeneratedTx2File.Close()

	// Sign using multisig. We're signing a tx on behalf of the multisig address,
	// even though the tx signer is NOT the multisig address. This is fine though,
	// as the main point of this test is to test the `--multisig` flag with an address
	// that is not in the keyring.
	_, err = authtestutil.TxSignExec(f.clientCtx, addr1, multiGeneratedTx2File.Name(), "--multisig", multisig)
	require.Contains(t, err.Error(), "error getting account from keybase")
}

func TestCLIMultisign(t *testing.T) {
	f := createTestSuite(t)

	// Generate 2 accounts and a multisig.
	account1, err := f.clientCtx.Keyring.Key("newAccount1")
	require.NoError(t, err)

	account2, err := f.clientCtx.Keyring.Key("newAccount2")
	require.NoError(t, err)

	multisigRecord, err := f.clientCtx.Keyring.Key("multi")
	require.NoError(t, err)

	addr, err := multisigRecord.GetAddress()
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   f.val.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 5)),
	}

	// Generate multisig transaction.
	multiGeneratedTx, err := clitestutil.SubmitTestTx(f.clientCtx, msgSend, addr, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	// Save tx to file
	multiGeneratedTxFile := testutil.WriteToNewTempFile(t, multiGeneratedTx.String())
	defer multiGeneratedTxFile.Close()

	addr1, err := account1.GetAddress()
	require.NoError(t, err)
	// Sign with account1
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)
	account1Signature, err := authtestutil.TxSignExec(f.clientCtx, addr1, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.NoError(t, err)

	sign1File := testutil.WriteToNewTempFile(t, account1Signature.String())
	defer sign1File.Close()

	addr2, err := account2.GetAddress()
	require.NoError(t, err)
	// Sign with account2
	account2Signature, err := authtestutil.TxSignExec(f.clientCtx, addr2, multiGeneratedTxFile.Name(), "--multisig", addr.String())
	require.NoError(t, err)

	sign2File := testutil.WriteToNewTempFile(t, account2Signature.String())
	defer sign2File.Close()

	f.clientCtx.Offline = false
	multiSigWith2Signatures, err := authtestutil.TxMultiSignExec(f.clientCtx, multisigRecord.Name, multiGeneratedTxFile.Name(), sign1File.Name(), sign2File.Name())
	require.NoError(t, err)

	// Write the output to disk
	signedTxFile := testutil.WriteToNewTempFile(t, multiSigWith2Signatures.String())
	defer signedTxFile.Close()

	_, err = authtestutil.TxValidateSignaturesExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)

	f.clientCtx.BroadcastMode = flags.BroadcastSync
	_, err = authtestutil.TxBroadcastExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)
}

func TestSignBatchMultisig(t *testing.T) {
	f := createTestSuite(t)

	// Fetch 2 accounts and a multisig.
	account1, err := f.clientCtx.Keyring.Key("newAccount1")
	require.NoError(t, err)
	account2, err := f.clientCtx.Keyring.Key("newAccount2")
	require.NoError(t, err)
	multisigRecord, err := f.clientCtx.Keyring.Key("multi")
	require.NoError(t, err)

	addr, err := multisigRecord.GetAddress()
	require.NoError(t, err)
	// Send coins from validator to multisig.
	sendTokens := sdk.NewInt64Coin("stake", 10)
	_, err = createBankMsg(
		f.clientCtx,
		f.val,
		addr,
		sdk.NewCoins(sendTokens),
		clitestutil.TestTxConfig{},
	)
	require.NoError(t, err)

	msgSend := &banktypes.MsgSend{
		FromAddress: addr.String(),
		ToAddress:   f.val.String(),
		Amount:      sdk.NewCoins(sdk.NewInt64Coin("stake", 5)),
	}

	generatedStd, err := clitestutil.SubmitTestTx(f.clientCtx, msgSend, addr, clitestutil.TestTxConfig{GenOnly: true})
	require.NoError(t, err)

	// Write the output to disk
	filename := testutil.WriteToNewTempFile(t, strings.Repeat(generatedStd.String(), 1))
	defer filename.Close()
	f.clientCtx.HomeDir = strings.Replace(f.clientCtx.HomeDir, "simd", "simcli", 1)

	addr1, err := account1.GetAddress()
	require.NoError(t, err)
	// sign-batch file
	res, err := authtestutil.TxSignBatchExec(f.clientCtx, addr1, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, f.clientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	require.NoError(t, err)
	require.Equal(t, 1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file
	file1 := testutil.WriteToNewTempFile(t, res.String())
	defer file1.Close()

	addr2, err := account2.GetAddress()
	require.NoError(t, err)
	// sign-batch file with account2
	res, err = authtestutil.TxSignBatchExec(f.clientCtx, addr2, filename.Name(), fmt.Sprintf("--%s=%s", flags.FlagChainID, f.clientCtx.ChainID), "--multisig", addr.String(), "--signature-only")
	require.NoError(t, err)
	require.Equal(t, 1, len(strings.Split(strings.Trim(res.String(), "\n"), "\n")))
	// write sigs to file2
	file2 := testutil.WriteToNewTempFile(t, res.String())
	defer file2.Close()
	_, err = authtestutil.TxMultiSignExec(f.clientCtx, multisigRecord.Name, filename.Name(), file1.Name(), file2.Name())
	require.NoError(t, err)
}

func TestGetBroadcastCommandOfflineFlag(t *testing.T) {
	cmd := authcli.GetBroadcastCommand()
	_ = testutil.ApplyMockIODiscardOutErr(cmd)
	cmd.SetArgs([]string{fmt.Sprintf("--%s=true", flags.FlagOffline), ""})

	require.EqualError(t, cmd.Execute(), "cannot broadcast tx during offline mode")
}

func TestGetBroadcastCommandWithoutOfflineFlag(t *testing.T) {
	f := createTestSuite(t)

	txCfg := f.clientCtx.TxConfig
	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxConfig(txCfg)

	ctx := context.Background()
	ctx = context.WithValue(ctx, client.ClientContextKey, &clientCtx)

	cmd := authcli.GetBroadcastCommand()
	_, out := testutil.ApplyMockIO(cmd)

	// Create new file with tx
	builder := txCfg.NewTxBuilder()
	builder.SetGasLimit(200000)

	err := builder.SetMsgs(banktypes.NewMsgSend("cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw", "cosmos1cxlt8kznps92fwu3j6npahx4mjfutydyene2qw", sdk.Coins{sdk.NewInt64Coin("stake", 10000)}))
	require.NoError(t, err)
	txContents, err := txCfg.TxJSONEncoder()(builder.GetTx())
	require.NoError(t, err)
	txFile := testutil.WriteToNewTempFile(t, string(txContents))
	defer txFile.Close()

	cmd.SetArgs([]string{txFile.Name()})
	err = cmd.ExecuteContext(ctx)
	require.Error(t, err)
	require.Contains(t, err.Error(), "connect: connection refused")
	require.Contains(t, out.String(), "connect: connection refused")
}

// TestTxWithoutPublicKey makes sure sending a proto tx message without the
// public key doesn't cause any error in the RPC layer (broadcast).
// See https://github.com/cosmos/cosmos-sdk/issues/7585 for more details.
func TestTxWithoutPublicKey(t *testing.T) {
	f := createTestSuite(t)

	txCfg := f.clientCtx.TxConfig

	valStr, err := f.clientCtx.AddressCodec.BytesToString(f.val)
	require.NoError(t, err)

	// Create a txBuilder with an unsigned tx.
	txBuilder := txCfg.NewTxBuilder()
	msg := banktypes.NewMsgSend(valStr, valStr, sdk.NewCoins(
		sdk.NewCoin("Stake", math.NewInt(10)),
	))
	err = txBuilder.SetMsgs(msg)
	require.NoError(t, err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("Stake", math.NewInt(150))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit())

	// Create a file with the unsigned tx.
	txJSON, err := txCfg.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	unsignedTxFile := testutil.WriteToNewTempFile(t, string(txJSON))
	defer unsignedTxFile.Close()

	// Sign the file with the unsignedTx.
	signedTx, err := authtestutil.TxSignExec(f.clientCtx, f.val, unsignedTxFile.Name(), fmt.Sprintf("--%s=true", cli.FlagOverwrite))
	require.NoError(t, err)

	// Remove the signerInfo's `public_key` field manually from the signedTx.
	// Note: this method is only used for test purposes! In general, one should
	// use txBuilder and TxEncoder/TxDecoder to manipulate txs.
	var tx tx.Tx
	err = f.clientCtx.Codec.UnmarshalJSON(signedTx.Bytes(), &tx)
	require.NoError(t, err)
	tx.AuthInfo.SignerInfos[0].PublicKey = nil
	// Re-encode the tx again, to another file.
	txJSON, err = f.clientCtx.Codec.MarshalJSON(&tx)
	require.NoError(t, err)
	signedTxFile := testutil.WriteToNewTempFile(t, string(txJSON))
	defer signedTxFile.Close()
	require.True(t, strings.Contains(string(txJSON), "\"public_key\":null"))

	// Broadcast tx, test that it shouldn't panic.
	f.clientCtx.BroadcastMode = flags.BroadcastSync
	out, err := authtestutil.TxBroadcastExec(f.clientCtx, signedTxFile.Name())
	require.NoError(t, err)
	var res sdk.TxResponse
	require.NoError(t, f.clientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	require.NotEqual(t, 0, res.Code)
}

// TestSignWithMultiSignersAminoJSON tests the case where a transaction with 2
// messages which has to be signed with 2 different keys. Sign and append the
// signatures using the CLI with Amino signing mode. Finally, send the
// transaction to the blockchain.
func TestSignWithMultiSignersAminoJSON(t *testing.T) {
	f := createTestSuite(t)

	val0, val1 := f.val, f.val1
	val0Coin := sdk.NewCoin("test1token", math.NewInt(10))
	val1Coin := sdk.NewCoin("test2token", math.NewInt(10))
	_, _, addr1 := testdata.KeyTestPubAddr()

	valStr, err := f.clientCtx.AddressCodec.BytesToString(val0)
	require.NoError(t, err)
	val1Str, err := f.clientCtx.AddressCodec.BytesToString(val1)
	require.NoError(t, err)

	addrStr, err := f.clientCtx.AddressCodec.BytesToString(addr1)
	require.NoError(t, err)
	// Creating a tx with 2 msgs from 2 signers: val0 and val1.
	// The validators need to sign with SIGN_MODE_LEGACY_AMINO_JSON,
	// because DIRECT doesn't support multi signers via the CLI.
	// Since we use amino, we don't need to pre-populate signer_infos.
	txBuilder := f.clientCtx.TxConfig.NewTxBuilder()
	err = txBuilder.SetMsgs(
		banktypes.NewMsgSend(valStr, addrStr, sdk.NewCoins(val0Coin)),
		banktypes.NewMsgSend(val1Str, addrStr, sdk.NewCoins(val1Coin)),
	)
	require.NoError(t, err)
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(10))))
	txBuilder.SetGasLimit(testdata.NewTestGasLimit() * 2)
	signers, err := txBuilder.GetTx().GetSigners()
	require.NoError(t, err)
	require.Equal(t, [][]byte{val0, val1}, signers)

	// Write the unsigned tx into a file.
	txJSON, err := f.clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	unsignedTxFile := testutil.WriteToNewTempFile(t, string(txJSON))
	defer unsignedTxFile.Close()

	// Let val0 sign first the file with the unsignedTx.
	signedByVal0, err := authtestutil.TxSignExec(f.clientCtx, val0, unsignedTxFile.Name(), "--overwrite", "--sign-mode=amino-json")
	require.NoError(t, err)
	signedByVal0File := testutil.WriteToNewTempFile(t, signedByVal0.String())
	defer signedByVal0File.Close()

	// Then let val1 sign the file with signedByVal0.
	val1AccNum, val1Seq, err := f.clientCtx.AccountRetriever.GetAccountNumberSequence(f.clientCtx, val1)
	require.NoError(t, err)

	signedTx, err := authtestutil.TxSignExec(
		f.clientCtx,
		val1,
		signedByVal0File.Name(),
		"--offline",
		fmt.Sprintf("--account-number=%d", val1AccNum),
		fmt.Sprintf("--sequence=%d", val1Seq),
		"--sign-mode=amino-json",
	)
	require.NoError(t, err)
	signedTxFile := testutil.WriteToNewTempFile(t, signedTx.String())
	defer signedTxFile.Close()

	res, err := authtestutil.TxBroadcastExec(
		f.clientCtx,
		signedTxFile.Name(),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
	)
	require.NoError(t, err)

	var txRes sdk.TxResponse
	require.NoError(t, f.clientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))
	require.Equal(t, uint32(0), txRes.Code, txRes.RawLog)
}

func TestAuxSigner(t *testing.T) {
	t.Skip("re-enable this when we bring back sign mode aux client testing")
	val0Coin := sdk.NewCoin("testtoken", math.NewInt(10))

	f := createTestSuite(t)

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
	}{
		{
			"error with SIGN_MODE_DIRECT_AUX and --aux unset",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
			},
			true,
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only set (ignores generate-only)",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
			},
			false,
		},
		{
			"no error with SIGN_MDOE_DIRECT_AUX mode and generate-only, tip flag set",
			[]string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagGenerateOnly),
				fmt.Sprintf("--%s=%s", flags.FlagTip, val0Coin.String()),
			},
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := govtestutil.MsgSubmitLegacyProposal(
				f.clientCtx,
				f.val.String(),
				"test",
				"test desc",
				govtypes.ProposalTypeText,
				tc.args...,
			)
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuxToFeeWithTips(t *testing.T) {
	// Skipping this test as it needs a simapp with the TipDecorator in post handler.
	t.Skip()

	f := createTestSuite(t)

	kb := f.clientCtx.Keyring
	acc, _, err := kb.NewMnemonic("tipperAccount", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)

	tipper, err := acc.GetAddress()
	require.NoError(t, err)
	tipperInitialBal := sdk.NewCoin("testtoken", math.NewInt(10000))

	feePayer := f.val
	fee := sdk.NewCoin("stake", math.NewInt(1000))
	tip := sdk.NewCoin("testtoken", math.NewInt(1000))

	_, err = createBankMsg(f.clientCtx, f.val, tipper, sdk.NewCoins(tipperInitialBal), clitestutil.TestTxConfig{})
	require.NoError(t, err)

	bal := getBalances(t, f.clientCtx, tipper, tip.Denom)
	require.True(t, bal.Equal(tipperInitialBal.Amount))

	testCases := []struct {
		name               string
		tipper             sdk.AccAddress
		feePayer           sdk.AccAddress
		tip                sdk.Coin
		expectErrAux       bool
		expectErrBroadCast bool
		errMsg             string
		tipperArgs         []string
		feePayerArgs       []string
	}{
		{
			name:     "when --aux and --sign-mode = direct set: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			expectErrAux: true,
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "both tipper, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "tipper uses DIRECT_AUX, fee payer uses AMINO: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "--tip flag unset: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "legacy amino json: no error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "tipper uses direct aux, fee payer uses direct: happy case",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
		},
		{
			name:     "chain-id mismatch: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      tip,
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			expectErrAux: false,
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
				fmt.Sprintf("--%s=%s", flags.FlagChainID, "foobar"),
			},
			expectErrBroadCast: true,
		},
		{
			name:     "wrong denom in tip: error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagTip, "1000wrongDenom"),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
				fmt.Sprintf("--%s=%s", flags.FlagFees, fee.String()),
			},
			errMsg: "insufficient funds",
		},
		{
			name:     "insufficient fees: less error",
			tipper:   tipper,
			feePayer: feePayer,
			tip:      sdk.Coin{Denom: "testtoken", Amount: math.NewInt(0)},
			tipperArgs: []string{
				fmt.Sprintf("--%s=%s", flags.FlagTip, tip),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirectAux),
				fmt.Sprintf("--%s=true", flags.FlagAux),
			},
			feePayerArgs: []string{
				fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
				fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeDirect),
				fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
				fmt.Sprintf("--%s=%s", flags.FlagFrom, feePayer),
			},
			errMsg: "insufficient fees",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := govtestutil.MsgSubmitLegacyProposal(
				f.clientCtx,
				tipper.String(),
				"test",
				"test desc",
				govtypes.ProposalTypeText,
				tc.tipperArgs...,
			)

			if tc.expectErrAux {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				genTxFile := testutil.WriteToNewTempFile(t, string(res.Bytes()))
				defer genTxFile.Close()

				switch {
				case tc.expectErrBroadCast:
					require.Error(t, err)

				case tc.errMsg != "":
					require.NoError(t, err)

					var txRes sdk.TxResponse
					require.NoError(t, f.clientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))
					require.Contains(t, txRes.RawLog, tc.errMsg)

				default:
					require.NoError(t, err)

					var txRes sdk.TxResponse
					require.NoError(t, f.clientCtx.Codec.UnmarshalJSON(res.Bytes(), &txRes))

					require.Equal(t, uint32(0), txRes.Code)
					require.NotNil(t, int64(0), txRes.Height)

					bal = getBalances(t, f.clientCtx, tipper, tc.tip.Denom)
					tipperInitialBal = tipperInitialBal.Sub(tc.tip)
					require.True(t, bal.Equal(tipperInitialBal.Amount))
				}
			}
		})
	}
}

func getBalances(t *testing.T, clientCtx client.Context, addr sdk.AccAddress, denom string) math.Int {
	t.Helper()

	resp, err := testutil.GetRequest(fmt.Sprintf("%s/cosmos/bank/v1beta1/balances/%s/by_denom?denom=%s", clientCtx.NodeURI, addr.String(), denom))
	require.NoError(t, err)

	var balRes banktypes.QueryAllBalancesResponse
	err = clientCtx.Codec.UnmarshalJSON(resp, &balRes)
	require.NoError(t, err)
	startTokens := balRes.Balances.AmountOf(denom)
	return startTokens
}

func createBankMsg(clientCtx client.Context, valAddr, toAddr sdk.AccAddress, amount sdk.Coins, cfg clitestutil.TestTxConfig) (testutil.BufferWriter, error) {
	msgSend := &banktypes.MsgSend{
		FromAddress: valAddr.String(),
		ToAddress:   toAddr.String(),
		Amount:      amount,
	}

	return clitestutil.SubmitTestTx(clientCtx, msgSend, valAddr, cfg)
}
