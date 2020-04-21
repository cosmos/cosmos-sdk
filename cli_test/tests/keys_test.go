package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/cli_test/helpers"
)

func TestCLIKeysAddMultisig(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// key names order does not matter
	f.KeysAdd("msig1", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBar, helpers.KeyBaz))
	ke1Address1 := f.KeysShow("msig1").Address
	f.KeysDelete("msig1")

	f.KeysAdd("msig2", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBaz, helpers.KeyBar))
	require.Equal(t, ke1Address1, f.KeysShow("msig2").Address)
	f.KeysDelete("msig2")

	f.KeysAdd("msig3", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBar, helpers.KeyBaz),
		"--nosort")
	f.KeysAdd("msig4", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBaz, helpers.KeyBar),
		"--nosort")
	require.NotEqual(t, f.KeysShow("msig3").Address, f.KeysShow("msig4").Address)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLISend(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	// Save key addresses for later use
	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)

	startTokens := sdk.TokensFromConsensusPower(50)
	require.Equal(t, startTokens, f.QueryBalances(fooAddr).AmountOf(denom))

	sendTokens := sdk.TokensFromConsensusPower(10)

	// It does not allow to send in offline mode
	success, _, stdErr := f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y", "--offline")
	require.Contains(t, stdErr, "no RPC client is defined in offline mode")
	require.False(f.T, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Send some tokens from one account to the other
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens, f.QueryBalances(barAddr).AmountOf(denom))
	require.Equal(t, startTokens.Sub(sendTokens), f.QueryBalances(fooAddr).AmountOf(denom))

	// Test --dry-run
	success, _, _ = f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--dry-run")
	require.True(t, success)

	// Test --generate-only
	success, stdout, stderr := f.TxSend(
		fooAddr.String(), barAddr, sdk.NewCoin(denom, sendTokens), "--generate-only=true",
	)
	require.Empty(t, stderr)
	require.True(t, success)
	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Len(t, msg.Msgs, 1)
	require.Len(t, msg.GetSignatures(), 0)

	// Check state didn't change
	require.Equal(t, startTokens.Sub(sendTokens), f.QueryBalances(fooAddr).AmountOf(denom))

	// test autosequencing
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(2), f.QueryBalances(barAddr).AmountOf(denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(2)), f.QueryBalances(fooAddr).AmountOf(denom))

	// test memo
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "--memo='testmemo'", "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure account balances match expected
	require.Equal(t, sendTokens.MulRaw(3), f.QueryBalances(barAddr).AmountOf(denom))
	require.Equal(t, startTokens.Sub(sendTokens.MulRaw(3)), f.QueryBalances(fooAddr).AmountOf(denom))

	f.Cleanup()
}

//-----------------------------------------------------------------------------------
//staking tx

func TestCLICreateValidator(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	barAddr := f.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens, f.QueryBalances(barAddr).AmountOf(denom))

	// Generate a create validator transaction and ensure correctness
	success, stdout, stderr := f.TxStakingCreateValidator(barAddr.String(), consPubKey, sdk.NewInt64Coin(denom, 2), "--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	newValTokens := sdk.TokensFromConsensusPower(2)
	success, _, _ = f.TxStakingCreateValidator(barAddr.String(), consPubKey, sdk.NewCoin(denom, newValTokens), "--dry-run")
	require.True(t, success)

	// Create the validator
	f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure funds were deducted properly
	require.Equal(t, sendTokens.Sub(newValTokens), f.QueryBalances(barAddr).AmountOf(denom))

	// Ensure that validator state is as expected
	validator := f.QueryStakingValidator(barVal)
	require.Equal(t, validator.OperatorAddress, barVal)
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	// Query delegations to the validator
	validatorDelegations := f.QueryStakingDelegationsTo(barVal)
	require.Len(t, validatorDelegations, 1)
	require.NotZero(t, validatorDelegations[0].Shares)

	// unbond a single share
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(1))
	success = f.TxStakingUnbond(keyBar, unbondAmt.String(), barVal, "-y")
	require.True(t, success)
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure bonded staking is correct
	remainingTokens := newValTokens.Sub(unbondAmt.Amount)
	validator = f.QueryStakingValidator(barVal)
	require.Equal(t, remainingTokens, validator.Tokens)

	// Get unbonding delegations from the validator
	validatorUbds := f.QueryStakingUnbondingDelegationsFrom(barVal)
	require.Len(t, validatorUbds, 1)
	require.Len(t, validatorUbds[0].Entries, 1)
	require.Equal(t, remainingTokens.String(), validatorUbds[0].Entries[0].Balance.String())

	f.Cleanup()
}

func TestCLIEditValidator(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	barAddr := f.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens, f.QueryBalances(barAddr).AmountOf(denom))

	newValTokens := sdk.TokensFromConsensusPower(2)

	// Create the validator
	f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure that validator state is as expected
	validator := f.QueryStakingValidator(barVal)
	require.Equal(t, validator.OperatorAddress, barVal)
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	// update moniker with test-edit
	var updateMoniker = "test-edit"
	if validator.Description.Moniker != "" {
		updateMoniker = validator.Description.Moniker + "-" + updateMoniker
	}

	// update details with test-details
	var updateDetails = "test-details"
	if validator.Description.Details != "" {
		updateDetails = validator.Description.Details + "-" + updateDetails
	}

	// update website with http://test-edit.co
	var updateWebsite = "http://test-edit.co"
	if validator.Description.Website != "" {
		updateWebsite = validator.Description.Website + "(or)" + updateWebsite
	}

	// Test --generate-only
	success, stdout, stderr := f.TxStakingEditValidator(barAddr.String(),
		fmt.Sprintf("--moniker=%s", updateMoniker),
		fmt.Sprintf("--details=%s", updateDetails),
		fmt.Sprintf("--website=%s", updateWebsite),
		"--generate-only")
	require.True(f.T, success)
	require.Empty(f.T, stderr)

	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxStakingEditValidator(barAddr.String(),
		fmt.Sprintf("--moniker=%s", updateMoniker),
		fmt.Sprintf("--details=%s", updateDetails),
		fmt.Sprintf("--website=%s", updateWebsite),
		"--dry-run")
	require.True(t, success)

	// Note: Commission cannot be changed more than once within 24 hrs
	// Edit validator's info
	success, _, err := f.TxStakingEditValidator(keyBar,
		fmt.Sprintf("--moniker=%s", updateMoniker),
		fmt.Sprintf("--details=%s", updateDetails),
		fmt.Sprintf("--website=%s", updateWebsite),
		"-y")

	require.Equal(t, success, true)
	require.Equal(t, err, "")
	tests.WaitForNextNBlocksTM(1, f.Port)

	updatedValidator := f.QueryStakingValidator(barVal)

	// Ensure validator's moniker got changed
	require.Equal(t, updatedValidator.Description.Moniker, updateMoniker)

	// Ensure validator's details got changed
	require.Equal(t, updatedValidator.Description.Details, updateDetails)

	// Ensure validator's website got changed
	require.Equal(t, updatedValidator.Description.Website, updateWebsite)

	f.Cleanup()
}

func TestCLIDelegate(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	fooAddr := f.KeyAddress(keyFoo)
	barAddr := f.KeyAddress(keyBar)
	barVal := sdk.ValAddress(barAddr)

	consPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	require.Equal(t, sendTokens, f.QueryBalances(barAddr).AmountOf(denom))

	newValTokens := sdk.TokensFromConsensusPower(2)

	// Create the validator
	f.TxStakingCreateValidator(keyBar, consPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure that validator state is as expected
	validator := f.QueryStakingValidator(barVal)
	require.Equal(t, validator.OperatorAddress, barVal)
	require.True(sdk.IntEq(t, newValTokens, validator.Tokens))

	delegateTokens := sdk.TokensFromConsensusPower(5)

	// Test --generate-only
	success, stdout, stderr := f.TxStakingDelegate(validator.OperatorAddress.String(), fooAddr.String(), sdk.NewCoin(denom, delegateTokens), "--generate-only")
	require.Equal(t, success, true)
	require.Empty(f.T, stderr)

	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	// Test --dry-run
	success, _, _ = f.TxStakingDelegate(validator.OperatorAddress.String(), fooAddr.String(), sdk.NewCoin(denom, delegateTokens), "--dry-run")
	require.Equal(t, success, true)

	// Start delegate tokens form keyfoo
	success, _, err := f.TxStakingDelegate(validator.OperatorAddress.String(), keyFoo, sdk.NewCoin(denom, delegateTokens), "-y")
	require.Equal(t, success, true)
	require.Equal(t, err, "")

	// Wait for the tx to commit into a block
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Read all delegations of a validator
	validatorDelegations := f.QueryStakingDelegationsTo(barVal)

	// Check the length, since the there are only 2 delegations length should be equal to 2
	require.Len(t, validatorDelegations, 2)
	delegatorAddress := f.KeysShow(keyFoo).Address
	var delegatedAccount staking.Delegation

	for i := 0; i < len(validatorDelegations); i++ {
		if validatorDelegations[i].DelegatorAddress.String() == delegatorAddress {
			delegatedAccount = validatorDelegations[i]
			break
		}
	}

	// Ensure the delegated amount should be greater than zero
	require.NotZero(t, delegatedAccount.Shares)

	// Ensure the amount equal to the delegated balance
	require.Equal(t, delegatedAccount.Shares, delegateTokens.ToDec())

	f.Cleanup()
}

func TestCLIRedelegate(t *testing.T) {
	t.Parallel()
	f := InitFixtures(t)

	// start gaiad server
	proc := f.GDStart()
	defer proc.Stop(false)

	// Create the 1st validator
	barAddr := f.KeyAddress(keyBar)
	srcValAddr := sdk.ValAddress(barAddr)

	srcPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())

	sendTokens := sdk.TokensFromConsensusPower(10)
	f.TxSend(keyFoo, barAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tokens sent to the dst address(i.e., barAddr)
	require.Equal(t, sendTokens, f.QueryBalances(barAddr).AmountOf(denom))

	newValTokens := sdk.TokensFromConsensusPower(2)

	f.TxStakingCreateValidator(keyBar, srcPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure that validator1 state is as expected
	srcVal := f.QueryStakingValidator(srcValAddr)
	require.Equal(t, srcVal.OperatorAddress, srcValAddr)
	require.True(sdk.IntEq(t, newValTokens, srcVal.Tokens))

	// Create the 2nd validator
	bazAddr := f.KeyAddress(keyBaz)
	dstValAddr := sdk.ValAddress(bazAddr)

	dstPubKey := sdk.MustBech32ifyPubKey(sdk.Bech32PubKeyTypeConsPub, ed25519.GenPrivKey().PubKey())
	f.TxSend(keyFoo, bazAddr, sdk.NewCoin(denom, sendTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure tokens sent to the dst address(i.e., bazAddr)
	require.Equal(t, sendTokens, f.QueryBalances(bazAddr).AmountOf(denom))

	success, _, err := f.TxStakingCreateValidator(keyBaz, dstPubKey, sdk.NewCoin(denom, newValTokens), "-y")
	tests.WaitForNextNBlocksTM(1, f.Port)

	// Ensure that validator2 state is as expected
	dstVal := f.QueryStakingValidator(dstValAddr)
	require.Equal(t, dstVal.OperatorAddress, dstValAddr)
	require.True(sdk.IntEq(t, newValTokens, dstVal.Tokens))

	redelegateValTokens := sdk.TokensFromConsensusPower(1)

	// Test --dry-run
	success, _, _ = f.TxStakingReDelegate(srcVal.OperatorAddress.String(), dstVal.OperatorAddress.String(),
		barAddr.String(), sdk.NewCoin(denom, redelegateValTokens), "--dry-run")
	require.Equal(t, success, true)

	// Test --generate-only
	success, stdout, stderr := f.TxStakingReDelegate(srcVal.OperatorAddress.String(), dstVal.OperatorAddress.String(),
		barAddr.String(), sdk.NewCoin(denom, redelegateValTokens), "--generate-only")
	require.Equal(t, success, true)
	require.Empty(f.T, stderr)

	msg := unmarshalStdTx(f.T, stdout)
	require.NotZero(t, msg.Fee.Gas)
	require.Equal(t, len(msg.Msgs), 1)
	require.Equal(t, 0, len(msg.GetSignatures()))

	success, _, err = f.TxStakingReDelegate(srcVal.OperatorAddress.String(), dstVal.OperatorAddress.String(),
		keyBar, sdk.NewCoin(denom, redelegateValTokens), "-y")

	// Ensure the redelegate tx succeed
	require.Equal(t, success, true)
	require.Equal(t, err, "")

	tests.WaitForNextNBlocksTM(1, f.Port)

	// Query validator's info after redelegate
	srcValDels := f.QueryStakingDelegationsTo(srcValAddr)
	dstValDels := f.QueryStakingDelegationsTo(dstValAddr)

	delegatedAccount := findDelegateAccount(dstValDels, f.KeysShow(keyBar).Address)
	// Ensure the delegated amount should be greater than zero
	require.NotZero(t, delegatedAccount.Shares)

	// Ensure the amount equal to the redelegated balance
	require.Equal(t, delegatedAccount.Shares, redelegateValTokens.ToDec())

	delegatedAccount = findDelegateAccount(srcValDels, f.KeysShow(keyBar).Address)
	// Ensure the delegated amount should be greater than zero
	require.NotZero(t, delegatedAccount.Shares)

	// Ensure the amount equal subtracted delegated balance
	require.Equal(t, delegatedAccount.Shares, newValTokens.Sub(redelegateValTokens).ToDec())
	f.Cleanup()
}
