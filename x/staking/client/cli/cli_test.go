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
	t.SkipNow() // Recreate when using CLI tests.

	t.Parallel()
	f := cli.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	barAddr := f.KeyAddress(cli.KeyBar)
	barVal := sdk.ValAddress(barAddr)

	// Check for the params
	params := testutil.QueryStakingParameters(f)
	require.NotEmpty(t, params)

	// Query for the staking pool
	pool := testutil.QueryStakingPool(f)
	require.NotEmpty(t, pool)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	bankclienttestutil.TxSend(f, cli.KeyFoo, barAddr, sdk.NewCoin(cli.Denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens.String(), bankclienttestutil.QueryBalances(f, barAddr).AmountOf(cli.Denom).String())

	// Generate a create validator transaction and ensure correctness
	success, stdout, stderr := testutil.TxStakingCreateValidator(f, barAddr.String(), consPubKey, sdk.NewInt64Coin(cli.Denom, 2), "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	msg := cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

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

	// Edit validator
	// params to be changed in edit validator (NOTE: a validator can only change its commission once per day)
	newMoniker := "test-moniker"
	newWebsite := "https://cosmos.network"
	newIdentity := "6A0D65E29A4CBC8D"
	newDetails := "To-infinity-and-beyond!"

	// Test --generate-only"
	success, stdout, stderr = testutil.TxStakingEditValidator(f, barAddr.String(), newMoniker, newWebsite, newIdentity, newDetails, "--generate-only")
	require.True(t, success)
	require.True(t, success)
	require.Empty(t, stderr)

	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	success, _, _ = testutil.TxStakingEditValidator(f, cli.KeyBar, newMoniker, newWebsite, newIdentity, newDetails, "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	udpatedValidator := testutil.QueryStakingValidator(f, barVal)
	require.Equal(t, udpatedValidator.Description.Moniker, newMoniker)
	require.Equal(t, udpatedValidator.Description.Identity, newIdentity)
	require.Equal(t, udpatedValidator.Description.Website, newWebsite)
	require.Equal(t, udpatedValidator.Description.Details, newDetails)

	// unbond a single share
	validators := testutil.QueryStakingValidators(f)
	require.Len(t, validators, 2)

	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(1))
	success = testutil.TxStakingUnbond(f, cli.KeyBar, unbondAmt.String(), barVal, "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded staking is correct
	remainingTokens := newValTokens.Sub(unbondAmt.Amount)
	validator = testutil.QueryStakingValidator(f, barVal)
	require.Equal(t, remainingTokens, validator.Tokens)

	// Query for historical info
	require.NotEmpty(t, testutil.QueryStakingHistoricalInfo(f, 1))

	// Get unbonding delegations from the validator
	validatorUbds := testutil.QueryStakingUnbondingDelegationsFrom(f, barVal)
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, remainingTokens.String(), validatorUbds[0].Entries[0].Balance.String())

	// Query staking unbonding delegation
	ubd := testutil.QueryStakingUnbondingDelegation(f, barAddr.String(), barVal.String())
	require.NotEmpty(t, ubd)

	// Query staking unbonding delegations
	ubds := testutil.QueryStakingUnbondingDelegations(f, barAddr.String())
	require.Len(t, ubds, 1)

	fooAddr := f.KeyAddress(cli.KeyFoo)

	delegateTokens := sdk.TokensFromConsensusPower(2)
	delegateAmount := sdk.NewCoin(cli.Denom, delegateTokens)

	// Delegate txn
	// Generate a create validator transaction and ensure correctness
	success, stdout, stderr = testutil.TxStakingDelegate(f, fooAddr.String(), barVal.String(), delegateAmount, "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	// Delegate
	success, _, err := testutil.TxStakingDelegate(f, cli.KeyFoo, barVal.String(), delegateAmount, "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)
	require.Empty(t, err)
	require.True(t, success)

	// Query the delegation from foo address to barval
	delegation := testutil.QueryStakingDelegation(f, fooAddr.String(), barVal)
	require.NotZero(t, delegation.Shares)

	// Query the delegations from foo address to barval
	delegations := testutil.QueryStakingDelegations(f, barAddr.String())
	require.Len(t, delegations, 1)

	fooVal := sdk.ValAddress(fooAddr)

	// Redelegate
	success, stdout, stderr = testutil.TxStakingRedelegate(f, fooAddr.String(), barVal.String(), fooVal.String(), delegateAmount, "--generate-only")
	require.True(t, success)
	require.Empty(t, stderr)

	msg = cli.UnmarshalStdTx(t, f.Cdc, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	success, _, err = testutil.TxStakingRedelegate(f, cli.KeyFoo, barVal.String(), fooVal.String(), delegateAmount, "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)
	require.Empty(t, err)
	require.True(t, success)

	redelegation := testutil.QueryStakingRedelegation(f, fooAddr.String(), barVal.String(), fooVal.String())
	require.Len(t, redelegation, 1)

	redelegations := testutil.QueryStakingRedelegations(f, fooAddr.String())
	require.Len(t, redelegations, 1)

	redelegationsFrom := testutil.QueryStakingRedelegationsFrom(f, barVal.String())
	require.Len(t, redelegationsFrom, 1)

	f.Cleanup()
}
