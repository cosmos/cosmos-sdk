// +build cli_test

package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankclienttestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/client/testutil"
)

func TestCLICreateValidator(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	barAddr := f.KeyAddress(cli.KeyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	bankclienttestutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens, bankclienttestutil.QueryBalances(f, barAddr).AmountOf(cli.Denom))

	//Generate a create validator transaction and ensure correctness
	success, stdout, stderr := testutil.TxStakingCreateValidator(f, barAddr.String(), consPubKey, sdk.NewInt64Coin(cli.Denom, 2), "--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg := cli.UnmarshalStdTx(f.T, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	newValTokens := sdk.TokensFromConsensusPower(2)
	success, _, _ = testutil.TxStakingCreateValidator(f, barAddr.String(), consPubKey, sdk.NewCoin(cli.Denom, newValTokens), "--dry-run")
	require.True(t, success)

	// Create the validator
	testutil.TxStakingCreateValidator(f, cli.KeyBar, consPubKey, sdk.NewCoin(cli.Denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure funds were deducted properly
	require.Equal(t, sendTokens.Sub(newValTokens), bankclienttestutil.QueryBalances(f, barAddr).AmountOf(cli.Denom))

	// Ensure that validator state is as expected
	validator := testutil.QueryStakingValidator(f, barVal)
	require.Equal(t, validator.OperatorAddress, barVal)
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := testutil.QueryStakingDelegationsTo(f, barVal)
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(1))
	success = testutil.TxStakingUnbond(f, cli.KeyBar, unbondAmt.String(), barVal, "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded staking is correct
	remainingTokens := newValTokens.Sub(unbondAmt.Amount)
	validator = testutil.QueryStakingValidator(f, barVal)
	require.Equal(t, remainingTokens, validator.Tokens)

	// Get unbonding delegations from the validator
	validatorUbds := testutil.QueryStakingUnbondingDelegationsFrom(f, barVal)
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, remainingTokens.String(), validatorUbds[0].Entries[0].Balance.String())

	f.Cleanup()
}
