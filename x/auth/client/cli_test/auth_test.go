// +build cli_test

package cli_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/tests"
	cli "github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli_test"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli_test"
	"github.com/stretchr/testify/require"
)

func TestCLIValidateSignatures(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server
	proc := f.SDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	// generate sendTx with default gas
	success, stdout, stderr := bankcli.TxSend(f, fooAddr.String(), barAddr, sdk.NewInt64Coin(cli.Denom, 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// write  unsigned tx to file
	unsignedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// validate we can successfully sign
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name())
	require.True(t, success)
	stdTx := cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// validate signatures
	success, _, _ = authcli.TxSign(f, cli.KeyFoo, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)

	// modify the transaction
	stdTx.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := cli.MarshalStdTx(t, f.Cdc, stdTx)
	modSignedTxFile := cli.WriteToNewTempFile(t, string(bz))
	defer os.Remove(modSignedTxFile.Name())

	// validate signature validation failure due to different transaction sig bytes
	success, _, _ = authcli.TxSign(f, cli.KeyFoo, modSignedTxFile.Name(), "--validate-signatures")
	require.False(t, success)

	f.Cleanup()
}

func TestCLISendGenerateSignAndBroadcast(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server
	proc := f.SDStart()
	defer proc.Stop(false)

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
	unsignedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--validate-signatures")
	require.False(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n\n", fooAddr.String()), stdout)

	// Test sign

	// Does not work in offline mode
	success, stdout, stderr = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--offline")
	require.Contains(t, stderr, "required flag(s) \"account-number\", \"sequence\" not set")
	require.False(t, success)

	// But works offline if we set account number and sequence
	success, _, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--offline", "--account-number", "1", "--sequence", "1")
	require.True(t, success)

	// Sign transaction
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name())
	require.True(t, success)
	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 1, len(msg.GetSignatures()))
	require.Equal(t, fooAddr.String(), msg.GetSigners()[0].String())

	// Write the output to disk
	signedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Test sign --validate-signatures
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)
	require.Equal(t, fmt.Sprintf("Signers:\n  0: %v\n\nSignatures:\n  0: %v\t\t\t[OK]\n\n", fooAddr.String(),
		fooAddr.String()), stdout)

	// Ensure foo has right amount of funds
	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, bankcli.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// Test broadcast

	// Does not work in offline mode
	success, _, stderr = authcli.TxBroadcast(f, signedTxFile.Name(), "--offline")
	require.Contains(t, stderr, "cannot broadcast tx during offline mode")
	require.False(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	success, stdout, _ = authcli.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account state
	require.Equal(t, sendTokens, bankcli.QueryBalances(f, barAddr).AmountOf(cli.Denom))
	require.Equal(t, startTokens.Sub(sendTokens), bankcli.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	f.Cleanup()
}

func TestCLIMultisignInsufficientCosigners(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.SDStart()
	defer proc.Stop(false)

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
	unsignedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Multisign, not enough signatures
	success, stdout, _ = authcli.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = authcli.TxSign(f, cli.KeyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.False(t, success)

	// Broadcast the transaction
	success, stdOut, _ := authcli.TxBroadcast(f, signedTxFile.Name())
	require.Contains(t, stdOut, "signature verification failed")
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIEncode(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server
	proc := f.SDStart()
	defer proc.Stop(false)

	// Build a testing transaction and write it to disk
	barAddr := f.KeyAddress(cli.KeyBar)
	keyAddr := f.KeyAddress(cli.KeyFoo)

	sendTokens := sdk.TokensFromConsensusPower(10)
	success, stdout, stderr := bankcli.TxSend(f, keyAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only", "--memo", "deadbeef")
	require.True(t, success)
	require.Empty(t, stderr)

	// Write it to disk
	jsonTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(jsonTxFile.Name())

	// Run the encode command, and trim the extras from the stdout capture
	success, base64Encoded, _ := authcli.TxEncode(f, jsonTxFile.Name())
	require.True(t, success)
	trimmedBase64 := strings.Trim(base64Encoded, "\"\n")

	// Decode the base64
	decodedBytes, err := base64.StdEncoding.DecodeString(trimmedBase64)
	require.Nil(t, err)

	// Check that the transaction decodes as epxceted
	var decodedTx auth.StdTx
	require.Nil(t, f.Cdc.UnmarshalBinaryBare(decodedBytes, &decodedTx))
	require.Equal(t, "deadbeef", decodedTx.Memo)
}

func TestCLIMultisignSortSignatures(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.SDStart()
	defer proc.Stop(false)

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
	unsignedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Sign with baz's key
	success, stdout, _ = authcli.TxSign(f, cli.KeyBaz, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String())
	require.True(t, success)

	// Write the output to disk
	bazSignatureFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(bazSignatureFile.Name())

	// Multisign, keys in different order
	success, stdout, _ = authcli.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		bazSignatureFile.Name(), fooSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = authcli.TxSign(f, cli.KeyFooBarBaz, signedTxFile.Name(), "--validate-signatures")
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = authcli.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestGaiaCLIMultisign(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start gaiad server with minimum fees
	proc := f.SDStart()
	defer proc.Stop(false)

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
	unsignedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(unsignedTxFile.Name())

	// Sign with foo's key
	success, stdout, _ = authcli.TxSign(f, cli.KeyFoo, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	fooSignatureFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(fooSignatureFile.Name())

	// Sign with bar's key
	success, stdout, _ = authcli.TxSign(f, cli.KeyBar, unsignedTxFile.Name(), "--multisig", fooBarBazAddr.String(), "-y")
	require.True(t, success)

	// Write the output to disk
	barSignatureFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(barSignatureFile.Name())

	// Multisign

	// Does not work in offline mode
	success, stdout, _ = authcli.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()}, "--offline")
	require.Contains(t, "couldn't verify signature", stdout)
	require.False(t, success)

	// Success multisign
	success, stdout, _ = authcli.TxMultisign(f, unsignedTxFile.Name(), cli.KeyFooBarBaz, []string{
		fooSignatureFile.Name(), barSignatureFile.Name()})
	require.True(t, success)

	// Write the output to disk
	signedTxFile := cli.WriteToNewTempFile(t, stdout)
	defer os.Remove(signedTxFile.Name())

	// Validate the multisignature
	success, _, _ = authcli.TxSign(f, cli.KeyFooBarBaz, signedTxFile.Name(), "--validate-signatures", "-y")
	require.True(t, success)

	// Broadcast the transaction
	success, _, _ = authcli.TxBroadcast(f, signedTxFile.Name())
	require.True(t, success)

	// Cleanup testing directories
	f.Cleanup()
}
