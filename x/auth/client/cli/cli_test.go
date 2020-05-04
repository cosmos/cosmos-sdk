// +build cli_test

package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/client/testutil"
)

func TestCLIValidateSignatures(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	f.ValidateGenesis()

	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	// generate sendTx with default gas
	success, stdout, stderr := testutil.TxSend(f, fooAddr.String(), barAddr, sdk.NewInt64Coin("stake", 10), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	// write  unsigned tx to file
	unsignedTxFile, cleanup := tests.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// validate we can successfully sign
	success, stdout, _ = testutil.TxSign(f, cli.KeyFoo, unsignedTxFile.Name())
	require.True(t, success)

	stdTx := cli.UnmarshalStdTx(t, f.Cdc, stdout)

	require.Equal(t, len(stdTx.Msgs), 1)
	require.Equal(t, 1, len(stdTx.GetSignatures()))
	require.Equal(t, fooAddr.String(), stdTx.GetSigners()[0].String())

	// write signed tx to file
	signedTxFile, cleanup := tests.WriteToNewTempFile(t, stdout)
	t.Cleanup(cleanup)

	// validate signatures
	success, _, _ = testutil.TxValidateSignatures(f, signedTxFile.Name())
	require.True(t, success)

	// modify the transaction
	stdTx.Memo = "MODIFIED-ORIGINAL-TX-BAD"
	bz := cli.MarshalStdTx(t, f.Cdc, stdTx)
	modSignedTxFile, cleanup := tests.WriteToNewTempFile(t, string(bz))
	t.Cleanup(cleanup)

	// validate signature validation failure due to different transaction sig bytes
	success, _, _ = testutil.TxValidateSignatures(f, modSignedTxFile.Name())
	require.False(t, success)

	// Cleanup testing directories
	f.Cleanup()
}
