// +build cli_test

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
)

func TestCLISend(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	// Save key addresses for later uspackage testse
	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, testutil.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	sendTokens := sdk.TokensFromConsensusPower(10)

	// It does not allow to send in offline mode
	success, _, stdErr := testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "-y", "--offline")
	require.Contains(t, stdErr, "no RPC client is defined in offline mode")
	require.False(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Send some tokens from one account to the other
	testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.String(), testutil.QueryBalances(f, barAddr).AmountOf(cli.Denom).String())
	require.Equal(t, startTokens.Sub(sendTokens).String(), testutil.QueryBalances(f, fooAddr).AmountOf(cli.Denom).String())

	// Test --dry-run
	success, _, _ = testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--dry-run")
	require.True(t, success)

	// Test --generate-only
	success, stdout, stderr := testutil.TxSend(
		f, fooAddr.String(), barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--generate-only=true",
	)
	require.Empty(t, stderr)
	require.True(t, success)
	msg := cli.UnmarshalStdTx(f.T, f.Cdc, stdout)
	t.Log(msg)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	// Check state didn't change
	require.Equal(t, startTokens.Sub(sendTokens), testutil.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// test autosequencing
	testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(2), testutil.QueryBalances(f, barAddr).AmountOf(cli.Denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), testutil.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	// test memo
	testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "--memo='testmemo'", "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(3), testutil.QueryBalances(f, barAddr).AmountOf(cli.Denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), testutil.QueryBalances(f, fooAddr).AmountOf(cli.Denom))

	f.Cleanup()
}

func TestCLIMinimumFees(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	fees := fmt.Sprintf(
		"--minimum-gas-prices=%s,%s",
		sdk.NewDecCoinFromDec(cli.FeeDenom, minGasPrice),
		sdk.NewDecCoinFromDec(cli.Fee2Denom, minGasPrice),
	)
	proc := f.SDStart(fees)
	t.Cleanup(func() { proc.Stop(false) })

	barAddr := f.KeyAddress(cli.KeyBar)

	// Send a transaction that will get rejected
	success, stdOut, _ := testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.Fee2Denom, 10), "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ correct fees pass
	txFees := fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(cli.FeeDenom, 2))
	success, _, _ = testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.Fee2Denom, 10), txFees, "-y")
	require.True(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tx w/ improper fees fails
	txFees = fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(cli.FeeDenom, 1))
	success, _, _ = testutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.FooDenom, 10), txFees, "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(f.T, success)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIGasPrices(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.SDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(cli.FeeDenom, minGasPrice)))
	t.Cleanup(func() { proc.Stop(false) })

	barAddr := f.KeyAddress(cli.KeyBar)

	// insufficient gas prices (tx fails)
	badGasPrice, _ := sdk.NewDecFromStr("0.000003")
	success, stdOut, _ := testutil.TxSend(
		f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.FooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(cli.FeeDenom, badGasPrice)), "-y")
	require.Contains(t, stdOut, "insufficient fees")
	require.True(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	// sufficient gas prices (tx passes)
	success, _, _ = testutil.TxSend(
		f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.FooDenom, 50),
		fmt.Sprintf("--gas-prices=%s", sdk.NewDecCoinFromDec(cli.FeeDenom, minGasPrice)), "-y")
	require.True(t, success)

	// wait for a block confirmation
	tests.WaitForNextNBlocksTM(1, f.Port)

	f.Cleanup()
}

func TestCLIFeesDeduction(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server with minimum fees
	minGasPrice, _ := sdk.NewDecFromStr("0.000006")
	proc := f.SDStart(fmt.Sprintf("--minimum-gas-prices=%s", sdk.NewDecCoinFromDec(cli.FeeDenom, minGasPrice)))
	t.Cleanup(func() { proc.Stop(false) })

	// Save key addresses for later use
	fooAddr := f.KeyAddress(cli.KeyFoo)
	barAddr := f.KeyAddress(cli.KeyBar)

	fooAmt := testutil.QueryBalances(f, fooAddr).AmountOf(cli.FooDenom)

	// test simulation
	success, _, _ := testutil.TxSend(
		f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.FooDenom, 1000),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(cli.FeeDenom, 2)), "--dry-run")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	require.Equal(t, fooAmt.Int64(), testutil.QueryBalances(f, fooAddr).AmountOf(cli.FooDenom).Int64())

	// insufficient funds (coins + fees) tx fails
	largeCoins := sdk.TokensFromConsensusPower(10000000)
	success, stdOut, _ := testutil.TxSend(
		f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.FooDenom, largeCoins),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(cli.FeeDenom, 2)), "-y")
	require.Contains(t, stdOut, "insufficient funds")
	require.True(t, success)

	// Wait for a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// ensure state didn't change
	require.Equal(t, fooAmt.Int64(), testutil.QueryBalances(f, fooAddr).AmountOf(cli.FooDenom).Int64())

	// test success (transfer = coins + fees)
	success, _, _ = testutil.TxSend(
		f, cli.KeyFoo, barAddr, sdk.NewInt64Coin(cli.FooDenom, 500),
		fmt.Sprintf("--fees=%s", sdk.NewInt64Coin(cli.FeeDenom, 2)), "-y")
	require.True(t, success)

	f.Cleanup()
}

func TestCLIQuerySupply(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	totalSupply := testutil.QueryTotalSupply(f)
	totalSupplyOf := testutil.QueryTotalSupplyOf(f, cli.FooDenom)

	require.Equal(t, cli.TotalCoins, totalSupply)
	require.True(sdk.IntEq(t, cli.TotalCoins.AmountOf(cli.FooDenom), totalSupplyOf))

	f.Cleanup()
}
