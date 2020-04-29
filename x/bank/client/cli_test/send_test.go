// +build cli_test

package cli_test

import (
	"fmt"
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

func TestCLIMinimumFees(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	fees := fmt.Sprintf(
		"--minimum-gas-prices=%s,%s",
		sdk.NewDecCoinFromDec(helpers.FeeDenom, minGasPrice),
		sdk.NewDecCoinFromDec(helpers.Fee2Denom, minGasPrice),
	)
	proc := f.SDStart(fees)
	defer proc.Stop(false)

	barAddr := f.KeyAddress(helpers.KeyBar)

	// Send a transaction that will get rejected
	success, stdOut, _ := bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.Fee2Denom, 10), "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ correct fees pass
	txFees := fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(helpers.FeeDenom, 2))
	success, _, _ = bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.Fee2Denom, 10), txFees, "-y")
	require.True(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ improper fees fails
	txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(helpers.FeeDenom, 1))
	success, _, _ = bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.FooDenom, 10), txFees, "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(f.T, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIGasPrices(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.SDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(helpers.FeeDenom, minGasPrice)))
	defer proc.Stop(false)

	barAddr := f.KeyAddress(helpers.KeyBar)

	// insufficient gas prices (tx fails)
	badGasPrice, _ := sdk.NewDecFromStr("0.000003")
	success, stdOut, _ := bankcli.TxSend(
		f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.FooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(helpers.FeeDenom, badGasPrice)), "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	// sufficient gas prices (tx passes)
	success, _, _ = bankcli.TxSend(
		f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.FooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(helpers.FeeDenom, minGasPrice)), "-y")
	require.True(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	f.Cleanup()
}

func TestCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.SDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(helpers.FeeDenom, minGasPrice)))
	defer proc.Stop(false)

	// Save key addresses for later use
	fooAddr := f.KeyAddress(helpers.KeyFoo)
	barAddr := f.KeyAddress(helpers.KeyBar)

	fooAmt := bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.FooDenom)

	// test simulation
	success, _, _ := bankcli.TxSend(
		f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.FooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(helpers.FeeDenom, 2)), "--dry-run")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	require.Equal(t, fooAmt.Int64(), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.FooDenom).Int64())

	// insufficient funds (coins + fees) tx fails
	largeCoins := sdk.TokensFromConsensusPower(10000000)
	success, stdOut, _ := bankcli.TxSend(
		f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.FooDenom, largeCoins),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(helpers.FeeDenom, 2)), "-y")
	require.Contains(t, stdOut, "insufficient funds")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	require.Equal(t, fooAmt.Int64(), bankcli.QueryBalances(f, fooAddr).AmountOf(helpers.FooDenom).Int64())

	// test success (transfer = coins + fees)
	success, _, _ = bankcli.TxSend(
		f, helpers.KeyFoo, barAddr, sdk.NewInt64Coin(helpers.FooDenom, 500),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(helpers.FeeDenom, 2)), "-y")
	require.True(t, success)

	f.Cleanup()
}
