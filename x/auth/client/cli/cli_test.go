// +build cli_test

package cli_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdktestutil "github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

func TestCLIValidateSignatures(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	f.ValidateGenesis()

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	// generate sendTx with default gas
	success, stdout, stderr := bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewInt64Coin("stake", 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// write  unsigned tx to file
	unsignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// validate we can successfully sign
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name())
	require.True(t, success)

	stdTx := cli.UnmarshalStdTx(t, f.Cdc, stdout)

	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// validate signatures
	success, _, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// modify the transaction
	stdTx.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := cli.MarshalStdTx(t, f.Cdc, stdTx)
	modSignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, string(bz))
	t.Cleanup(cleanup)

	// validate signature validation failure due to different transaction sig bytes
	success, _, _ = testutil.TxValidateSignatures(f, modSignedTxFile.Name())
	require.False(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLISignBatch(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	sendTokens := sdk.TokensFromConsensusPower(10)
	success, generatedStdTx, stderr := bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only")

	require.True(t, success)
	require.Empty(t, stderr)

	// Write the output to disk
	batchfile, cleanup1 := sdktestutil.WriteToNewTempFile(t, strings.Repeat(generatedStdTx, 3))
	t.Cleanup(cleanup1)

	// sign-batch file - offline is set but account-number and sequence are not
	success, _, stderr = testutil.TxSignBatch(f, cli.KeyFoo, batchfile.Name(), "--offline")
	require.Contains(t, stderr, "required flag(s) \"account-number\", \"sequence\" not set")
	require.False(t, success)

	// sign-batch file
	success, stdout, stderr := testutil.TxSignBatch(f, cli.KeyFoo, batchfile.Name())
	require.True(t, success)
	require.Empty(t, stderr)
	require.Equal(t, 3, len(strings.Split(strings.Trim(stdout, "\n"), "\n")))

	// sign-batch file
	success, stdout, stderr = testutil.TxSignBatch(f, cli.KeyFoo, batchfile.Name(), "--signature-only")
	require.True(t, success)
	require.Empty(t, stderr)
	require.Equal(t, 3, len(strings.Split(strings.Trim(stdout, "\n"), "\n")))

	malformedFile, cleanup2 := sdktestutil.WriteToNewTempFile(t, fmt.Sprintf("%smalformed", generatedStdTx))
	t.Cleanup(cleanup2)

	// sign-batch file
	success, stdout, stderr = testutil.TxSignBatch(f, cli.KeyFoo, malformedFile.Name())
	require.False(t, success)
	require.Equal(t, 1, len(strings.Split(strings.Trim(stdout, "\n"), "\n")))
	require.Equal(t, "ERROR: cannot parse disfix JSON wrapper: invalid character 'm' looking for beginning of value\n", stderr)

	// sign-batch file
	success, stdout, _ = testutil.TxSignBatch(f, cli.KeyFoo, malformedFile.Name(), "--signature-only")
	require.False(t, success)
	require.Equal(t, 1, len(strings.Split(strings.Trim(stdout, "\n"), "\n")))

	f.Cleanup()
}

func TestCLISendGenerateSignAndBroadcast(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	// Test generate sendTx with default gas
	sendTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr := bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg := cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(flags.DefaultGasLimit))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx with --gas=$amount
	success, stdout, stderr = bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--gas=100", "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, msg.Fee.Gas, uint64(100))
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test generate sendTx, estimate gas
	success, stdout, stderr = bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)
	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.True(t, msg.Fee.Gas > 0)
	require.Equal(t, len(msg.Msgs), 1)

	// Write the output to disk
	unsignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Test validate-signatures
	success, stdout, _ = testutil.TxValidateSignatures(f, unsignedTxFile.Name())
	require.False(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", fooAddr.String()), stdout)

	// Test sign

	// Does not work in offline mode
	success, stdout, stderr = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--offline")
	require.Contains(t, stderr, "required flag(s) \"account-number\", \"sequence\" not set")
	require.False(t, success)

	// But works offline if we set account number and sequence
	success, _, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	require.True(t, success)

	// Sign transaction
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name())
	require.True(t, success)
	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Test validate-signatures
	success, stdout, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n  0: %v\t\t\t[OK]\n\n", fooAddr.String(),
		fooAddr.String()), stdout)

	// Ensure foo has right amount of funds
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, bankcli.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// Test broadcast

	// Does not work in offline mode
	success, _, stderr = testutil.TxBroadcast(f, signedTxFile.Name(), "--offline")
	require.Contains(t, stderr, "cannot broadcast tx during offline mode")
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	success, stdout, _ = testutil.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account state
	require.Equal(t, sendTokens, bankcli.QueryBalances(f, barAddr).AmountOf(cli.Denom))
	require.Equal(t, startTokens.Sub(sendTokens), bankcli.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	f.Cleanup()
}

func TestCLIMultisignInsufficientCosigners(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooBarBazAddr := f.KeyAddress(cli.KeyFooBarBaz)
	barAddr := f.KeyAddress(cli.KeyBar)

	// Send some tokens from one account to the other
	success, _, _ := bankcli.TxSend(f, cli.KeyFoo, fooBarBazAddr, sdk.NewInt64Coin(cli.Denom, 10), "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Test generate sendTx with multisig
	success, stdout, _ := bankcli.TxSend(f, fooBarBazAddr.String(), barAddr, sdk.NewInt64Coin(cli.Denom, 5), "--generate-only")
	require.True(t, success)

	// Write the output to disk
	unsignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with foo's key
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Multisign, not enough signatures
	success, stdout, _ = testutil.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Validate the multisignature
	success, _, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.False(t, success)

	// Broadcast the transaction
	success, stdOut, _ := testutil.TxBroadcast(f, signedTxFile.Name())
	require.Contains(t, stdOut, "signature verification failed")
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIEncode(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	// Build a testing transaction and write it to disk
	barAddr := f.KeyAddress(cli.KeyBar)
	keyAddr := f.KeyAddress(cli.KeyFoo)

	sendTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr := bankcli.TxSend(f, keyAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only", "--memo", "deadbeef")
	require.True(t, success)
	require.Empty(t, stderr)

	// Write it to disk
	jsonTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Run the encode command
	success, base64Encoded, _ := testutil.TxEncode(f, jsonTxFile.Name())
	require.True(t, success)
	trimmedBase64 := strings.Trim(base64Encoded, "\"\n")
	// Check that the transaction decodes as expected
	success, stdout, stderr = testutil.TxDecode(f, trimmedBase64)
	decodedTx := cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, "deadbeef", decodedTx.Memo)
}

func TestCLIMultisignSortSignatures(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooBarBazAddr := f.KeyAddress(cli.KeyFooBarBaz)
	barAddr := f.KeyAddress(cli.KeyBar)

	// Send some tokens from one account to the other
	success, _, _ := bankcli.TxSend(f, cli.KeyFoo, fooBarBazAddr, sdk.NewInt64Coin(cli.Denom, 10), "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, int64(10), bankcli.QueryBalances(f, fooBarBazAddr).AmountOf(cli.Denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, _ := bankcli.TxSend(f, fooBarBazAddr.String(), barAddr, sdk.NewInt64Coin(cli.Denom, 5), "--generate-only")
	require.True(t, success)

	// Write the output to disk
	unsignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with foo's key
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with baz's key
	success, stdout, _ = testutil.TxSign(f, cli.KeyBaz, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	bazSignatureFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Multisign, keys in different order
	success, stdout, _ = testutil.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		bazSignatureFile.Name(), fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Validate the multisignature
	success, _, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = testutil.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIMultisign(t *testing.T) {
	t.SkipNow()
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	fooBarBazAddr := f.KeyAddress(cli.KeyFooBarBaz)
	bazAddr := f.KeyAddress(cli.KeyBaz)

	// Send some tokens from one account to the other
	success, _, _ := bankcli.TxSend(f, cli.KeyFoo, fooBarBazAddr, sdk.NewInt64Coin(cli.Denom, 10), "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, int64(10), bankcli.QueryBalances(f, fooBarBazAddr).AmountOf(cli.Denom).Int64())

	// Test generate sendTx with multisig
	success, stdout, stderr := bankcli.TxSend(f, fooBarBazAddr.String(), bazAddr, sdk.NewInt64Coin(cli.Denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// Write the output to disk
	unsignedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with foo's key
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Sign with bar's key
	success, stdout, _ = testutil.TxSign(f, cli.KeyBar, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	barSignatureFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Multisign

	// Does not work in offline mode
	success, stdout, _ = testutil.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()}, "--offline")
	require.Contains(t, "couldn't verify signature", stdout)
	require.False(t, success)

	// Success multisign
	success, stdout, _ = testutil.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile, cleanup := sdktestutil.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// Validate the multisignature
	success, _, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = testutil.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}
