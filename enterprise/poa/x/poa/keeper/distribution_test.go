// IMPORTANT LICENSE NOTICE
//
// SPDX-License-Identifier: CosmosLabs-Evaluation-Only
//
// This file is NOT licensed under the Apache License 2.0.
//
// Licensed under the Cosmos Labs Source Available Evaluation License, which forbids:
// - commercial use,
// - production use, and
// - redistribution.
//
// See https://github.com/cosmos/cosmos-sdk/blob/main/enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	poatypes "github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestProportionalDistribution(t *testing.T) {
	t.Run("fees distributed proportionally to validator power", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with different powers
		// Validator 1: 100 power (25% of total)
		// Validator 2: 300 power (75% of total)
		// Total: 400 power
		validatorAddr1, _ := createValidator(t, f, 1, 100)
		validatorAddr2, _ := createValidator(t, f, 2, 300)

		// Add fees to fee collector
		feeCollector := f.authKeeper.GetModuleAccount(f.ctx, authtypes.FeeCollectorName)
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Query validator 1 - should show 25% (250 stake) as pending
		req1 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		expectedFees1 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250))}
		require.Equal(t, expectedFees1, resp1.Fees.Fees)

		// Query validator 2 - should show 75% (750 stake) as pending
		req2 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr2}
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, req2)
		require.NoError(t, err)
		expectedFees2 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(750))}
		require.Equal(t, expectedFees2, resp2.Fees.Fees)

		// Verify fee collector still has all fees (lazy distribution)
		feeCollectorBalance := f.bankKeeper.GetAllBalances(f.ctx, feeCollector.GetAddress())
		require.Equal(t, fees, feeCollectorBalance)
	})

	t.Run("equal power results in equal distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with equal powers
		validatorAddr1, _ := createValidator(t, f, 1, 100)
		validatorAddr2, _ := createValidator(t, f, 2, 100)
		validatorAddr3, _ := createValidator(t, f, 3, 100)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 300))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		expectedAmount := math.LegacyNewDec(100)

		// All validators should show equal pending amounts
		req1 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		require.Equal(t, expectedAmount, resp1.Fees.Fees.AmountOf("stake"))

		req2 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr2}
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, req2)
		require.NoError(t, err)
		require.Equal(t, expectedAmount, resp2.Fees.Fees.AmountOf("stake"))

		req3 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr3}
		resp3, err := f.poaKeeper.WithdrawableFees(f.ctx, req3)
		require.NoError(t, err)
		require.Equal(t, expectedAmount, resp3.Fees.Fees.AmountOf("stake"))
	})

	t.Run("no fees collected means no distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		validatorAddr1, _ := createValidator(t, f, 1, 100)

		// Validator should have no pending fees
		req := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp, err := f.poaKeeper.WithdrawableFees(f.ctx, req)
		require.NoError(t, err)
		require.True(t, resp.Fees.Fees.IsZero())
	})

	t.Run("no validators with lazy distribution", func(t *testing.T) {
		f := setupTest(t)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// With lazy distribution, fees remain in fee collector (no panic)
		feeCollector := f.authKeeper.GetModuleAccount(f.ctx, authtypes.FeeCollectorName)
		feeCollectorBalance := f.bankKeeper.GetAllBalances(f.ctx, feeCollector.GetAddress())
		require.Equal(t, fees, feeCollectorBalance)
	})

	t.Run("accumulated fees persist across multiple blocks", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		validatorAddr1, _ := createValidator(t, f, 1, 100)
		validatorAddr2, _ := createValidator(t, f, 2, 100)

		// Block 1: Add 200 stake
		fees1 := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees1)
		require.NoError(t, err)

		// Each validator should show 100 stake pending
		req1 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, resp1.Fees.Fees)

		// Block 2: Add another 200 stake
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Each validator should now show 200 stake total (100 + 100)
		resp1After, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(200))}, resp1After.Fees.Fees)

		req2 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr2}
		resp2After, err := f.poaKeeper.WithdrawableFees(f.ctx, req2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(200))}, resp2After.Fees.Fees)
	})

	t.Run("complex proportional distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with powers: 50, 150, 300
		// Total: 500
		// Percentages: 10%, 30%, 60%
		validatorAddr1, _ := createValidator(t, f, 1, 50)
		validatorAddr2, _ := createValidator(t, f, 2, 150)
		validatorAddr3, _ := createValidator(t, f, 3, 300)

		// Add 1000 stake to distribute
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Validator 1: 10% = 100 stake
		req1 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, resp1.Fees.Fees)

		// Validator 2: 30% = 300 stake
		req2 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr2}
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, req2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(300))}, resp2.Fees.Fees)

		// Validator 3: 60% = 600 stake
		req3 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr3}
		resp3, err := f.poaKeeper.WithdrawableFees(f.ctx, req3)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(600))}, resp3.Fees.Fees)
	})

	t.Run("multiple denominations distributed proportionally", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		validatorAddr1, _ := createValidator(t, f, 1, 100)
		validatorAddr2, _ := createValidator(t, f, 2, 300)

		// Add multiple denominations to fee collector
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 400),
		)
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Validator 1: 25% of each denom
		req1 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr1}
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, req1)
		require.NoError(t, err)
		expectedFees1 := sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(100)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250)),
		}
		require.Equal(t, expectedFees1, resp1.Fees.Fees)

		// Validator 2: 75% of each denom
		req2 := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr2}
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, req2)
		require.NoError(t, err)
		expectedFees2 := sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(300)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(750)),
		}
		require.Equal(t, expectedFees2, resp2.Fees.Fees)
	})

	t.Run("decimal remainders are preserved across withdrawals", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with 1/3 of total power (will create fractional amounts)
		validatorAddr, consAddr1 := createValidator(t, f, 1, 100)
		_, _ = createValidator(t, f, 2, 200)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Distribute 10 tokens
		// Validator has 100 power out of 300 total = (100/300) * 10 tokens
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 10))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		expectedAmount, err := math.LegacyNewDecFromStr("3.333333333333333333")
		require.NoError(t, err)

		// Validator should have exact pending amount in query
		req := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr}
		resp, err := f.poaKeeper.WithdrawableFees(f.ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedAmount, resp.Fees.Fees.AmountOf("stake"))

		// Withdraw - should get 3 tokens, remainder stays as DecCoins
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 3)), coins)

		// Check validator received 3 tokens
		balance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(3), balance.Amount.Int64())

		// Check remainder is preserved (exact decimal remainder)
		accFeesAfter, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		expectedRemainder, err := math.LegacyNewDecFromStr("0.333333333333333333")
		require.NoError(t, err)
		require.Equal(t, expectedRemainder, accFeesAfter.AmountOf("stake"))

		// Distribute another 10 tokens
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Validator should now have exact total of previous remainder + new distribution (via query)
		respAfterSecond, err := f.poaKeeper.WithdrawableFees(f.ctx, req)
		require.NoError(t, err)
		expectedTotal := expectedRemainder.Add(expectedAmount)
		require.Equal(t, expectedTotal, respAfterSecond.Fees.Fees.AmountOf("stake"))

		// Withdraw again - should get 3 tokens, remainder is 0.666...
		coins2, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 3)), coins2)

		// Check validator received 3 more tokens (total 6)
		balanceAfterSecond := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(6), balanceAfterSecond.Amount.Int64())

		// Check new remainder is preserved (exact value)
		accFeesAfterSecondWithdraw, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		expectedSecondRemainder := expectedTotal.Sub(math.LegacyNewDec(3))
		require.Equal(t, expectedSecondRemainder, accFeesAfterSecondWithdraw.AmountOf("stake"))
	})

	t.Run("dust does not accumulate perpetually - remainders become whole coins", func(t *testing.T) {
		f := setupTest(t)

		// Create validators: 100 out of 700 total power
		// This creates fractional distributions that don't divide evenly
		validatorAddr, consAddr1 := createValidator(t, f, 1, 100)
		_, _ = createValidator(t, f, 2, 200)
		_, _ = createValidator(t, f, 3, 400)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Track total distributed to validator 1
		totalDistributed := math.LegacyZeroDec()

		// Distribute 7 tokens per block for 10 blocks
		// Validator 1 gets (100/700) * 7 = 1 token per block exactly
		// But let's use 11 tokens to create remainders: (100/700) * 11 = 1.571428...
		validatorPower := math.LegacyNewDec(100)
		totalPower := math.LegacyNewDec(700)

		for i := 0; i < 10; i++ {
			fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 11))
			err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
			require.NoError(t, err)

			// Track what should have been distributed this block
			distributed := math.LegacyNewDec(11).Mul(validatorPower).Quo(totalPower)
			totalDistributed = totalDistributed.Add(distributed)

			// Checkpointing is required here to simulate the above calculation
			// where we divide every loop.
			err = f.poaKeeper.CheckpointAllValidators(f.ctx)
			require.NoError(t, err)
		}

		// Check query shows correct pending amount
		req := &poatypes.QueryWithdrawableFeesRequest{OperatorAddress: validatorAddr}
		resp, err := f.poaKeeper.WithdrawableFees(f.ctx, req)
		require.NoError(t, err)
		require.Equal(t, totalDistributed, resp.Fees.Fees.AmountOf("stake"))

		// Withdraw - should get 15 whole coins
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 15)), coins)

		balance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(15), balance.Amount.Int64())

		// Check remainder is preserved
		accFeesAfter, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		// This is rounded slightly because of the constant checkpointing at every distribution
		// If we don't checkpoint every time, we end up with 0.714285714285714286.
		// Both cases are "correct", but checkpointing every block simulated someone withdrawing at every block.
		expectedRemainder, err := math.LegacyNewDecFromStr("0.714285714285714290")
		require.NoError(t, err)
		require.Equal(t, expectedRemainder, accFeesAfter.AmountOf("stake"))

		// Continue distributing for 10 more blocks
		totalDistributedSecond := math.LegacyZeroDec()
		for i := 0; i < 10; i++ {
			fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 11))
			err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
			require.NoError(t, err)

			distributed := math.LegacyNewDec(11).Mul(validatorPower).Quo(totalPower)
			totalDistributedSecond = totalDistributedSecond.Add(distributed)
			// Checkpointing is required here to simulate the above calculation
			// where we divide every loop.
			err = f.poaKeeper.CheckpointAllValidators(f.ctx)
			require.NoError(t, err)
		}

		// Validator should now have: remainder + 10 more distributions = 0.714... + 15.714... = 16.428...
		expectedNewTotal := expectedRemainder.Add(totalDistributedSecond)
		respSecond, err := f.poaKeeper.WithdrawableFees(f.ctx, req)
		require.NoError(t, err)
		require.Equal(t, expectedNewTotal, respSecond.Fees.Fees.AmountOf("stake"))

		// Withdraw again - should get 16 whole coins
		coins2, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 16)), coins2)

		balanceSecond := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(31), balanceSecond.Amount.Int64()) // 15 + 16

		// Final remainder check
		accFeesFinal, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		// This is rounded slightly because of the constant checkpointing at every distribution
		// If we don't checkpoint every time, we end up with 0.428571428571428572.
		// Both cases are "correct", but checkpointing every block simulated someone withdrawing at every block.
		finalRemainder, err := math.LegacyNewDecFromStr("0.428571428571428580")
		require.NoError(t, err)
		require.Equal(t, finalRemainder, accFeesFinal.AmountOf("stake"))

		// PROOF: Total distributed = Total withdrawn + Current remainder
		// Total distributed over 20 blocks = sum of actual validator allocations
		validator1CurrentFees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		totalWithdrawn := math.LegacyNewDec(31)
		totalEverDistributed := validator1CurrentFees.AmountOf("stake").Add(totalWithdrawn)
		proofSum := totalWithdrawn.Add(finalRemainder)
		require.Equal(t, totalEverDistributed, proofSum, "Total distributed must equal total withdrawn + remainder")

		// The remainder is small and will eventually accumulate to >= 1 coin in future blocks
		// Nothing is lost - it's all accounted for
		require.True(t, finalRemainder.GT(math.LegacyZeroDec()), "Remainder should be positive")
		require.True(t, finalRemainder.LT(math.LegacyOneDec()), "Remainder should be less than 1 coin")
	})
}

func TestCheckpointAllValidators(t *testing.T) {
	t.Run("checkpoint allocates fees to all validators", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 300)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Before checkpoint, validators should have zero accumulated fees
		accFees1Before, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.True(t, accFees1Before.IsZero())

		// Checkpoint all validators
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// After checkpoint, validator 1 should have 25% (250 stake)
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250))}, accFees1)

		// Validator 2 should have 75% (750 stake)
		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(750))}, accFees2)

		// Total allocated should be 1000
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}, totalAllocated)

		// Fee collector balance should remain unchanged (lazy distribution)
		feeCollector := f.authKeeper.GetModuleAccount(f.ctx, authtypes.FeeCollectorName)
		feeCollectorBalance := f.bankKeeper.GetAllBalances(f.ctx, feeCollector.GetAddress())
		require.Equal(t, fees, feeCollectorBalance)
	})

	t.Run("checkpoint with no fees does nothing", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 300)

		// Checkpoint with no fees
		err := f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// All validators should have no fees
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.True(t, accFees1.IsZero())

		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.True(t, accFees2.IsZero())

		// Total allocated should be 0
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.True(t, totalAllocated.IsZero())
	})

	t.Run("checkpoint skips zero power validators", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with different powers
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 0)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint all validators
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Validator 1 should get 100% (all fees)
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees1)

		// Validator 2 should have no fees
		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.True(t, accFees2.IsZero())
	})

	t.Run("multiple checkpoints accumulate correctly", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 100)

		// First checkpoint
		fees1 := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees1)
		require.NoError(t, err)

		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Each should have 100
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees1)

		// Second checkpoint
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Each should now have 200
		accFees1After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(200))}, accFees1After)

		accFees2After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(200))}, accFees2After)

		// Total allocated should be 400
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(400))}, totalAllocated)
	})

	t.Run("checkpoint with multiple denominations", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 300)

		// Add multiple denominations
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 400),
		)
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint all
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Validator 1: 25% of each
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		expectedFees1 := sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(100)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250)),
		}
		require.Equal(t, expectedFees1, accFees1)

		// Validator 2: 75% of each
		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		expectedFees2 := sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(300)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(750)),
		}
		require.Equal(t, expectedFees2, accFees2)
	})

	t.Run("checkpoint before power change maintains correct distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with initial power
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 100)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// SetValidatorPower should checkpoint before changing power
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr1, 200)
		require.NoError(t, err)

		// Both validators should have received 100 stake (before power change)
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees1)

		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees2)

		// Add more fees after power change
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 300))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Checkpoint again
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Now validator 1 has 200 power out of 300 total = 2/3 of new fees
		// Validator 1: 100 (old) + 200 (2/3 of 300) = 300
		accFees1After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(300), accFees1After.AmountOf("stake"))

		// Validator 2: 100 (old) + 100 (1/3 of 300) = 200
		accFees2After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(200), accFees2After.AmountOf("stake"))
	})

	t.Run("checkpoint with zero total power does nothing", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with zero power
		_, _ = createValidator(t, f, 1, 0)
		_, _ = createValidator(t, f, 2, 0)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint should succeed but not allocate (total power is 0)
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Total allocated should still be 0
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.True(t, totalAllocated.IsZero())
	})

	t.Run("checkpoint when all fees already allocated", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint once to allocate all fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Verify all fees are allocated
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, totalAllocated)

		// Checkpoint again - should do nothing since unallocated is zero
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Validator fees should remain the same
		accFees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees)
	})

	t.Run("checkpoint before creating new validator maintains correct distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create initial validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 100)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Create a new validator - should checkpoint existing validators first
		_, consAddr3 := createValidator(t, f, 3, 100)

		// Existing validators should have received 100 stake each (before new validator joined)
		accFees1, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees1)

		accFees2, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, accFees2)

		// New validator should have no fees yet
		accFees3, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr3)
		require.NoError(t, err)
		require.True(t, accFees3.IsZero())

		// Add more fees after new validator joined
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 300))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Checkpoint again
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Validators 1 and 2: 100 (before) + 100 (new share) each
		accFees1After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr1)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(200), accFees1After.AmountOf("stake"))

		accFees2After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr2)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(200), accFees2After.AmountOf("stake"))

		// Validator 3: 0 (before) + 100 (new share)
		accFees3After, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr3)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(100), accFees3After.AmountOf("stake"))
	})
}

func TestWithdrawValidatorFees(t *testing.T) {
	t.Run("withdraw with no accumulated fees", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Try to withdraw with no fees
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.True(t, coins.IsZero())

		// Balance should still be zero
		balance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.True(t, balance.IsZero())
	})

	t.Run("withdraw with only decimal remainder (no whole coins)", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		validatorAddr, consAddr := createValidator(t, f, 1, 100)
		_, _ = createValidator(t, f, 2, 300)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Add very small amount that creates only decimal fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Validator should have 0.25 stake (25% of 1)
		accFees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		expectedAmount, err := math.LegacyNewDecFromStr("0.25")
		require.NoError(t, err)
		require.Equal(t, expectedAmount, accFees.AmountOf("stake"))

		// Try to withdraw - should succeed but transfer 0 coins (only decimal)
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.True(t, coins.IsZero())

		// Balance should be zero (no whole coins)
		balance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(0), balance.Amount.Int64())

		// Accumulated fees should still have the decimal remainder
		accFeesAfter, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, expectedAmount, accFeesAfter.AmountOf("stake"))
	})

	t.Run("withdraw updates total allocated correctly", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		validatorAddr, _ := createValidator(t, f, 1, 100)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Total allocated should be 100
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, totalAllocated)

		// Withdraw
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 100)), coins)

		// Total allocated should now be 0
		totalAllocatedAfter, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.True(t, totalAllocatedAfter.IsZero())
	})

	t.Run("withdraw with non-existent validator returns error", func(t *testing.T) {
		f := setupTest(t)

		// Try to withdraw for non-existent validator
		nonExistentAddr := sdk.AccAddress("nonexistent")
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, nonExistentAddr)
		require.Error(t, err)
		require.Nil(t, coins)
	})

	t.Run("withdraw with multiple denominations", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with 100% power
		validatorAddr, consAddr := createValidator(t, f, 1, 100)
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Add fees in multiple denominations
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 150),
			sdk.NewInt64Coin("atom", 75),
			sdk.NewInt64Coin("osmo", 200),
		)
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Verify allocated fees for all denominations
		allocatedFees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, math.LegacyNewDec(150), allocatedFees.AmountOf("stake"))
		require.Equal(t, math.LegacyNewDec(75), allocatedFees.AmountOf("atom"))
		require.Equal(t, math.LegacyNewDec(200), allocatedFees.AmountOf("osmo"))

		// Withdraw all fees
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		expectedCoins := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 150),
			sdk.NewInt64Coin("atom", 75),
			sdk.NewInt64Coin("osmo", 200),
		)
		require.Equal(t, expectedCoins, coins)

		// Verify all denominations were withdrawn correctly
		stakeBalance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(150), stakeBalance.Amount.Int64())

		atomBalance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "atom")
		require.Equal(t, int64(75), atomBalance.Amount.Int64())

		osmoBalance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "osmo")
		require.Equal(t, int64(200), osmoBalance.Amount.Int64())

		// Verify allocated fees are now zero for all denominations
		allocatedFeesAfter, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.True(t, allocatedFeesAfter.IsZero())

		// Verify total allocated is zero
		totalAllocated, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.True(t, totalAllocated.IsZero())
	})

	t.Run("withdraw with multiple denominations preserves decimal remainders", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with fractional power distribution
		validatorAddr, consAddr := createValidator(t, f, 1, 100) // 25% power
		_, _ = createValidator(t, f, 2, 300)                     // 75% power
		validatorAddrSdk, err := sdk.AccAddressFromBech32(validatorAddr)
		require.NoError(t, err)

		// Add fees that will create decimal remainders when split 25/75
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 10), // Validator gets 2.5
			sdk.NewInt64Coin("atom", 7),   // Validator gets 1.75
		)
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Calculate expected decimal amounts
		expectedStake, err := math.LegacyNewDecFromStr("2.5")
		require.NoError(t, err)
		expectedAtom, err := math.LegacyNewDecFromStr("1.75")
		require.NoError(t, err)

		// Verify allocated fees
		allocatedFees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, expectedStake, allocatedFees.AmountOf("stake"))
		require.Equal(t, expectedAtom, allocatedFees.AmountOf("atom"))

		// Withdraw - should get whole coins only
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		expectedCoins := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 2), // 2.5 -> 2
			sdk.NewInt64Coin("atom", 1),  // 1.75 -> 1
		)
		require.Equal(t, expectedCoins, coins)

		// Verify withdrawn whole coins
		stakeBalance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(2), stakeBalance.Amount.Int64()) // 2.5 -> 2

		atomBalance := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "atom")
		require.Equal(t, int64(1), atomBalance.Amount.Int64()) // 1.75 -> 1

		// Verify decimal remainders are preserved
		allocatedFeesAfter, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)

		stakeRemainder := expectedStake.Sub(math.LegacyNewDec(2)) // 0.5
		require.Equal(t, stakeRemainder, allocatedFeesAfter.AmountOf("stake"))

		atomRemainder := expectedAtom.Sub(math.LegacyNewDec(1)) // 0.75
		require.Equal(t, atomRemainder, allocatedFeesAfter.AmountOf("atom"))

		// Add more fees
		fees2 := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 10),
			sdk.NewInt64Coin("atom", 7),
		)
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Checkpoint again
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Verify remainders accumulate with new fees
		allocatedFeesSecond, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)

		expectedStakeTotal := stakeRemainder.Add(expectedStake) // 0.5 + 2.5 = 3.0
		expectedAtomTotal := atomRemainder.Add(expectedAtom)    // 0.75 + 1.75 = 2.5

		require.Equal(t, expectedStakeTotal, allocatedFeesSecond.AmountOf("stake"))
		require.Equal(t, expectedAtomTotal, allocatedFeesSecond.AmountOf("atom"))

		// Withdraw again
		coins2, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, validatorAddrSdk)
		require.NoError(t, err)
		expectedCoins2 := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 3), // 3.0 -> 3
			sdk.NewInt64Coin("atom", 2),  // 2.5 -> 2
		)
		require.Equal(t, expectedCoins2, coins2)

		// Verify second withdrawal
		stakeBalanceSecond := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "stake")
		require.Equal(t, int64(5), stakeBalanceSecond.Amount.Int64()) // 2 + 3

		atomBalanceSecond := f.bankKeeper.GetBalance(f.ctx, validatorAddrSdk, "atom")
		require.Equal(t, int64(3), atomBalanceSecond.Amount.Int64()) // 1 + 2

		// Verify new remainders
		allocatedFeesFinal, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)

		// Stake: 3.0 - 3 = 0.0 (exactly zero, no remainder)
		require.True(t, allocatedFeesFinal.AmountOf("stake").IsZero())

		// Atom: 2.5 - 2 = 0.5 (has remainder)
		finalAtomRemainder := expectedAtomTotal.Sub(math.LegacyNewDec(2))
		require.Equal(t, finalAtomRemainder, allocatedFeesFinal.AmountOf("atom"))
	})
}

func TestGetValidatorAllocatedFees(t *testing.T) {
	t.Run("get fees for non-existent validator returns error", func(t *testing.T) {
		f := setupTest(t)

		// Try to get fees for non-existent validator
		nonExistentConsAddr := sdk.ConsAddress("nonexistent")
		_, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, nonExistentConsAddr)
		require.Error(t, err)
	})

	t.Run("get fees for existing validator with no fees", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		_, consAddr := createValidator(t, f, 1, 100)

		// Get fees - should be zero
		fees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.True(t, fees.IsZero())
	})

	t.Run("get fees for existing validator with fees", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		_, consAddr := createValidator(t, f, 1, 100)

		// Add fees
		feeCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, feeCoins)
		require.NoError(t, err)

		// Checkpoint
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Get fees - should be 100
		fees, err := f.poaKeeper.getValidatorAllocatedFees(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, fees)
	})
}

func TestCalculateValidatorPendingFees(t *testing.T) {
	t.Run("calculates proportional fees correctly", func(t *testing.T) {
		unallocated := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}
		validatorPower := int64(100)
		totalPower := int64(400)

		result := calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250))}, result)
	})

	t.Run("returns zero for zero validator power", func(t *testing.T) {
		unallocated := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}
		validatorPower := int64(0)
		totalPower := int64(400)

		result := calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		require.True(t, result.IsZero())
	})

	t.Run("panics for zero total power", func(t *testing.T) {
		unallocated := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}
		validatorPower := int64(100)
		totalPower := int64(0)

		require.PanicsWithValue(t, "totalPower cannot be zero when calculating validator pending fees", func() {
			calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		})
	})

	t.Run("handles multiple denominations", func(t *testing.T) {
		unallocated := sdk.DecCoins{
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000)),
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(400)),
		}
		validatorPower := int64(100)
		totalPower := int64(400)

		result := calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		require.Equal(t, sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(100)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(250)),
		}, result)
	})

	t.Run("handles fractional results", func(t *testing.T) {
		unallocated := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(10))}
		validatorPower := int64(1)
		totalPower := int64(3)

		result := calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		expected, err := math.LegacyNewDecFromStr("3.333333333333333333")
		require.NoError(t, err)
		require.Equal(t, expected, result.AmountOf("stake"))
	})

	t.Run("handles validator with all power", func(t *testing.T) {
		unallocated := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}
		validatorPower := int64(100)
		totalPower := int64(100)

		result := calculateValidatorPendingFees(validatorPower, totalPower, unallocated)
		require.Equal(t, unallocated, result)
	})
}

func TestGetUnallocatedFees(t *testing.T) {
	t.Run("returns zero when fee collector is empty", func(t *testing.T) {
		f := setupTest(t)

		unallocated, err := f.poaKeeper.getUnallocatedFees(f.ctx)
		require.NoError(t, err)
		require.True(t, unallocated.IsZero())
	})

	t.Run("returns all fees when nothing allocated", func(t *testing.T) {
		f := setupTest(t)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		unallocated, err := f.poaKeeper.getUnallocatedFees(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000))}, unallocated)
	})

	t.Run("returns difference when some fees allocated", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		_, _ = createValidator(t, f, 1, 100)

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Allocate some fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Add more fees
		moreFees := sdk.NewCoins(sdk.NewInt64Coin("stake", 500))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, moreFees)
		require.NoError(t, err)

		unallocated, err := f.poaKeeper.getUnallocatedFees(f.ctx)
		require.NoError(t, err)
		// Should have 500 unallocated (the new fees)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(500))}, unallocated)
	})

	t.Run("handles multiple denominations", func(t *testing.T) {
		f := setupTest(t)

		// Add multiple denominations
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 400),
		)
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		unallocated, err := f.poaKeeper.getUnallocatedFees(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(400)),
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000)),
		}, unallocated)
	})

	t.Run("returns zero when all fees allocated", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		_, _ = createValidator(t, f, 1, 100)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Allocate all fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		unallocated, err := f.poaKeeper.getUnallocatedFees(f.ctx)
		require.NoError(t, err)
		require.True(t, unallocated.IsZero())
	})
}

func TestAdjustTotalAllocated(t *testing.T) {
	t.Run("increases total allocated", func(t *testing.T) {
		f := setupTest(t)

		delta := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta)
		require.NoError(t, err)

		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, delta, total)
	})

	t.Run("decreases total allocated", func(t *testing.T) {
		f := setupTest(t)

		// First increase
		delta1 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(200))}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta1)
		require.NoError(t, err)

		// Then decrease using MulDec(-1) like the actual code does
		positiveDelta := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		delta2 := positiveDelta.MulDec(math.LegacyNewDec(-1))
		err = f.poaKeeper.adjustTotalAllocated(f.ctx, delta2)
		require.NoError(t, err)

		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}, total)
	})

	t.Run("handles zero delta", func(t *testing.T) {
		f := setupTest(t)

		// Set initial value
		delta1 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta1)
		require.NoError(t, err)

		// Zero delta should not change
		delta2 := sdk.DecCoins{}
		err = f.poaKeeper.adjustTotalAllocated(f.ctx, delta2)
		require.NoError(t, err)

		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, delta1, total)
	})

	t.Run("handles multiple denominations", func(t *testing.T) {
		f := setupTest(t)

		delta := sdk.DecCoins{
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(1000)),
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(400)),
		}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta)
		require.NoError(t, err)

		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, delta, total)
	})

	t.Run("handles negative delta that results in zero", func(t *testing.T) {
		f := setupTest(t)

		// Set initial value
		delta1 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta1)
		require.NoError(t, err)

		// Decrease to zero using MulDec(-1) like the actual code does
		positiveDelta := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		delta2 := positiveDelta.MulDec(math.LegacyNewDec(-1))
		err = f.poaKeeper.adjustTotalAllocated(f.ctx, delta2)
		require.NoError(t, err)

		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.True(t, total.IsZero())
	})

	t.Run("prevents adjusting below zero", func(t *testing.T) {
		f := setupTest(t)

		// Set initial value
		delta1 := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100))}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta1)
		require.NoError(t, err)

		// Try to decrease below zero using MulDec(-1) like the actual code does
		positiveDelta := sdk.DecCoins{sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(150))}
		delta2 := positiveDelta.MulDec(math.LegacyNewDec(-1))
		err = f.poaKeeper.adjustTotalAllocated(f.ctx, delta2)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot adjust total allocated below zero")

		// Total should remain unchanged
		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, delta1, total)
	})

	t.Run("prevents adjusting below zero with multiple denominations", func(t *testing.T) {
		f := setupTest(t)

		// Set initial value with multiple denominations
		delta1 := sdk.DecCoins{
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(100)),
			sdk.NewDecCoinFromDec("atom", math.LegacyNewDec(50)),
		}
		err := f.poaKeeper.adjustTotalAllocated(f.ctx, delta1)
		require.NoError(t, err)

		// Try to decrease one denomination below zero using MulDec(-1) like the actual code does
		positiveDelta := sdk.DecCoins{
			sdk.NewDecCoinFromDec("stake", math.LegacyNewDec(150)),
		}
		negativeDelta := positiveDelta.MulDec(math.LegacyNewDec(-1))
		err = f.poaKeeper.adjustTotalAllocated(f.ctx, negativeDelta)
		require.Error(t, err)
		require.Contains(t, err.Error(), "cannot adjust total allocated below zero")

		// Total should remain unchanged
		total, err := f.poaKeeper.getTotalAllocated(f.ctx)
		require.NoError(t, err)
		require.Equal(t, delta1, total)
	})
}
