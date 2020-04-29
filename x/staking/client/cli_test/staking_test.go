// +build cli_test

package cli_test

import (
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankcli "github.com/cosmos/cosmos-sdk/x/bank/client/cli_test"
	stakingcli "github.com/cosmos/cosmos-sdk/x/staking/client/cli_test"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"testing"
)

func TestCLICreateValidator(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	defer proc.Stop(false)

	barAddr := f.KeyAddress(helpers.KeyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	bankcli.TxSend(f, helpers.KeyFoo, barAddr, sdk.NewCoin(helpers.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens, bankcli.QueryBalances(f, barAddr).AmountOf(helpers.Denom))

	//Generate a create validator transaction and ensure correctness
	success, stdout, stderr := stakingcli.TxStakingCreateValidator(f, barAddr.String(), consPubKey, sdk.NewInt64Coin(helpers.Denom, 2), "--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg := helpers.UnmarshalStdTx(f.T, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	newValTokens := sdk.TokensFromConsensusPower(2)
	success, _, _ = stakingcli.TxStakingCreateValidator(f, barAddr.String(), consPubKey, sdk.NewCoin(helpers.Denom, newValTokens), "--dry-run")
	require.True(t, success)

	// Create the validator
	stakingcli.TxStakingCreateValidator(f, helpers.KeyBar, consPubKey, sdk.NewCoin(helpers.Denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure funds were deducted properly
	require.Equal(t, sendTokens.Sub(newValTokens), bankcli.QueryBalances(f, barAddr).AmountOf(helpers.Denom))

	// Ensure that validator state is as expected
	validator := stakingcli.QueryStakingValidator(f, barVal)
	require.Equal(t, validator.OperatorAddress, barVal)
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := stakingcli.QueryStakingDelegationsTo(f, barVal)
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(1))
	success = stakingcli.TxStakingUnbond(f, helpers.KeyBar, unbondAmt.String(), barVal, "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded staking is correct
	remainingTokens := newValTokens.Sub(unbondAmt.Amount)
	validator = stakingcli.QueryStakingValidator(f, barVal)
	require.Equal(t, remainingTokens, validator.Tokens)

	// Get unbonding delegations from the validator
	validatorUbds := stakingcli.QueryStakingUnbondingDelegationsFrom(f, barVal)
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, remainingTokens.String(), validatorUbds[0].Entries[0].Balance.String())

	f.Cleanup()
}
