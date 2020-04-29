// +build cli_test

package cli_test

import (
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli_test"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCLISend(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	defer proc.Stop(false)

	// Save key addresses for later uspackage testse
	fooAddr := f.KeyAddress(helpers.KeyFoo)
	barAddr := f.KeyAddress(helpers.KeyBar)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	sendTokens := sdk.TokensFromConsensusPower(10)

	// It does not allow to send in offline mode
	success, _, stdErr := bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "-y", "--offline")
	require.Contains(t, stdErr, "no RPC client is defined in offline mode")
	require.False(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Send some tokens from one account to the other
	bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens, bankcli.QueryBalances(f, barAddr).AmountOf(helpers.Denom))
	require.Equal(t, startTokens.Sub(sendTokens), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	// Test --dry-run
	success, _, _ = bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "--dry-run")
	require.True(t, success)

	// Test --generate-only
	success, stdout, stderr := bankcli.TxSend(
		f, fooAddr.String(), barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "--generate-only=true",
	)
	require.Empty(t, stderr)
	require.True(t, success)
	msg := helpers.UnmarshalStdTx(f.T, f.Cdc, stdout)
	t.Log(msg)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	// Check state didn't change
	require.Equal(t, startTokens.Sub(sendTokens), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	// test autosequencing
	bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(2), bankcli.QueryBalances(f, barAddr).AmountOf(helpers.Denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	// test memo
	bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "--memo='testmemo'", "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(3), bankcli.QueryBalances(f, barAddr).AmountOf(helpers.Denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.Denom))

	f.Cleanup()
}
