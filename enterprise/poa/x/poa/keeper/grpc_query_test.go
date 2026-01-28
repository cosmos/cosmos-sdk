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
// See ./enterprise/poa/LICENSE for full terms.
// Copyright (c) 2026 Cosmos Labs US Inc.

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

func TestValidatorsQueryDescendingOrder(t *testing.T) {
	f := setupTest(t)

	// Create validators with different power levels
	// We'll create them in a non-sorted order to verify the query sorts them
	opAddr1, _ := createValidator(t, f, 1, 100) // Medium power
	opAddr2, _ := createValidator(t, f, 2, 500) // Highest power
	opAddr3, _ := createValidator(t, f, 3, 50)  // Low power
	opAddr4, _ := createValidator(t, f, 4, 200) // Medium-high power
	opAddr5, _ := createValidator(t, f, 5, 0)   // No power

	// Query all validators
	resp, err := f.poaKeeper.Validators(f.ctx, &types.QueryValidatorsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Validators, 5)

	// Verify validators are returned in descending power order
	require.Equal(t, int64(500), resp.Validators[0].Power)
	require.Equal(t, opAddr2, resp.Validators[0].Metadata.OperatorAddress)

	require.Equal(t, int64(200), resp.Validators[1].Power)
	require.Equal(t, opAddr4, resp.Validators[1].Metadata.OperatorAddress)

	require.Equal(t, int64(100), resp.Validators[2].Power)
	require.Equal(t, opAddr1, resp.Validators[2].Metadata.OperatorAddress)

	require.Equal(t, int64(50), resp.Validators[3].Power)
	require.Equal(t, opAddr3, resp.Validators[3].Metadata.OperatorAddress)

	require.Equal(t, int64(0), resp.Validators[4].Power)
	require.Equal(t, opAddr5, resp.Validators[4].Metadata.OperatorAddress)

	// Verify that the list is strictly descending
	for i := 0; i < len(resp.Validators)-1; i++ {
		require.GreaterOrEqual(t, resp.Validators[i].Power, resp.Validators[i+1].Power,
			"Validators should be in descending power order")
	}
}

func TestValidatorsQueryPagination(t *testing.T) {
	f := setupTest(t)

	// Create 10 validators with different powers
	numValidators := 10
	operatorAddrs := make([]string, numValidators)
	powers := []int64{1000, 900, 800, 700, 600, 500, 400, 300, 200, 100}

	for i := 0; i < numValidators; i++ {
		opAddr, _ := createValidator(t, f, i+1, powers[i])
		operatorAddrs[i] = opAddr
	}

	t.Run("pagination with limit", func(t *testing.T) {
		// Request first page with limit of 3
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit: 3,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Len(t, resp.Validators, 3)
		require.NotNil(t, resp.Pagination)

		// Verify first page has highest power validators in descending order
		require.Equal(t, int64(1000), resp.Validators[0].Power)
		require.Equal(t, int64(900), resp.Validators[1].Power)
		require.Equal(t, int64(800), resp.Validators[2].Power)

		// Verify pagination response has next key
		require.NotNil(t, resp.Pagination.NextKey)
		require.True(t, len(resp.Pagination.NextKey) > 0)
	})

	t.Run("pagination with limit and key", func(t *testing.T) {
		// Get first page
		req1 := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit: 3,
			},
		}
		resp1, err := f.poaKeeper.Validators(f.ctx, req1)
		require.NoError(t, err)
		require.NotNil(t, resp1.Pagination.NextKey)

		// Get second page using next key
		req2 := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Key:   resp1.Pagination.NextKey,
				Limit: 3,
			},
		}
		resp2, err := f.poaKeeper.Validators(f.ctx, req2)
		require.NoError(t, err)
		require.Len(t, resp2.Validators, 3)

		// Verify second page continues from where first page left off
		require.Equal(t, int64(700), resp2.Validators[0].Power)
		require.Equal(t, int64(600), resp2.Validators[1].Power)
		require.Equal(t, int64(500), resp2.Validators[2].Power)

		// Verify no overlap between pages
		firstPagePowers := make(map[int64]bool)
		for _, v := range resp1.Validators {
			firstPagePowers[v.Power] = true
		}
		for _, v := range resp2.Validators {
			require.False(t, firstPagePowers[v.Power], "validator with power %d should not appear in both pages", v.Power)
		}
	})

	t.Run("pagination with count total", func(t *testing.T) {
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit:      5,
				CountTotal: true,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp.Pagination)
		require.Equal(t, uint64(numValidators), resp.Pagination.Total)
		require.Len(t, resp.Validators, 5)
	})

	t.Run("pagination with offset", func(t *testing.T) {
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Offset: 3,
				Limit:  3,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Validators, 3)

		// Verify we get validators starting from offset 3 (powers 700, 600, 500)
		require.Equal(t, int64(700), resp.Validators[0].Power)
		require.Equal(t, int64(600), resp.Validators[1].Power)
		require.Equal(t, int64(500), resp.Validators[2].Power)
	})

	t.Run("pagination maintains descending order", func(t *testing.T) {
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit: 5,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)

		// Verify validators are still in descending order within the page
		for i := 0; i < len(resp.Validators)-1; i++ {
			require.GreaterOrEqual(t, resp.Validators[i].Power, resp.Validators[i+1].Power,
				"validators should be in descending power order within paginated results")
		}
	})

	t.Run("pagination with limit exceeding total", func(t *testing.T) {
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit: 100, // More than total validators
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Validators, numValidators)
		// NextKey should be nil when all results are returned
		require.Nil(t, resp.Pagination.NextKey)
	})

	t.Run("pagination with reverse=false still returns descending order", func(t *testing.T) {
		// Even if reverse is explicitly set to false, we default to descending
		// (This tests our default behavior)
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Limit:   5,
				Reverse: false,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Validators, 5)

		// Should still be in descending order (our default)
		require.Equal(t, int64(1000), resp.Validators[0].Power)
		require.Equal(t, int64(900), resp.Validators[1].Power)
	})
}

func TestMultiPageValidatorsQuery(t *testing.T) {
	f := setupTest(t)

	// Create 15 validators with different powers
	numValidators := 15
	expectedPowers := make([]int64, numValidators)
	for i := 0; i < numValidators; i++ {
		power := int64((numValidators - i) * 50) // Powers: 750, 700, 650, ..., 50
		expectedPowers[i] = power
		createValidator(t, f, i+1, power)
	}

	// Paginate through all validators with page size of 4
	pageSize := uint64(4)
	allValidators := make([]types.Validator, 0)
	seenPowers := make(map[int64]bool)
	var nextKey []byte
	pageNum := 0

	for {
		req := &types.QueryValidatorsRequest{
			Pagination: &query.PageRequest{
				Key:   nextKey,
				Limit: pageSize,
			},
		}

		resp, err := f.poaKeeper.Validators(f.ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)

		// Verify page has expected number of results (except possibly last page)
		if resp.Pagination.NextKey != nil {
			require.Len(t, resp.Validators, int(pageSize), "page %d should have %d validators", pageNum, pageSize)
		} else {
			// Last page may have fewer results
			require.LessOrEqual(t, len(resp.Validators), int(pageSize), "last page should have at most %d validators", pageSize)
		}

		// Verify validators are in descending order within the page
		for i := 0; i < len(resp.Validators)-1; i++ {
			require.GreaterOrEqual(t, resp.Validators[i].Power, resp.Validators[i+1].Power,
				"page %d: validators should be in descending order", pageNum)
		}

		// Collect validators and check for duplicates
		for _, v := range resp.Validators {
			require.False(t, seenPowers[v.Power], "duplicate validator with power %d found on page %d", v.Power, pageNum)
			seenPowers[v.Power] = true
			allValidators = append(allValidators, v)
		}

		// Check if we've reached the last page
		if len(resp.Pagination.NextKey) == 0 {
			break
		}

		nextKey = resp.Pagination.NextKey
		pageNum++
	}

	// Verify we retrieved all validators
	require.Len(t, allValidators, numValidators, "should have retrieved all %d validators", numValidators)

	// Verify all validators are in descending order overall
	for i := 0; i < len(allValidators)-1; i++ {
		require.GreaterOrEqual(t, allValidators[i].Power, allValidators[i+1].Power,
			"all validators should be in descending order across pages")
	}

	// Verify we got all expected powers
	for _, expectedPower := range expectedPowers {
		found := false
		for _, v := range allValidators {
			if v.Power == expectedPower {
				found = true
				break
			}
		}
		require.True(t, found, "validator with power %d should be present", expectedPower)
	}

	// Verify we had multiple pages
	require.Greater(t, pageNum, 0, "should have had multiple pages")
	expectedMinPages := numValidators/int(pageSize) - 1
	require.Greater(t, pageNum, expectedMinPages, "should have had at least %d pages", expectedMinPages)
}

func TestTotalPowerQuery(t *testing.T) {
	t.Run("returns zero when no validators exist", func(t *testing.T) {
		f := setupTest(t)

		resp, err := f.poaKeeper.TotalPower(f.ctx, &types.QueryTotalPowerRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(0), resp.TotalPower)
	})

	t.Run("returns correct total power with single validator", func(t *testing.T) {
		f := setupTest(t)

		// Create one validator with power 100
		createValidator(t, f, 1, 100)

		resp, err := f.poaKeeper.TotalPower(f.ctx, &types.QueryTotalPowerRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(100), resp.TotalPower)
	})

	t.Run("returns correct total power with multiple validators", func(t *testing.T) {
		f := setupTest(t)

		// Create validators with different powers
		_, consAddr1 := createValidator(t, f, 1, 100)
		_, consAddr2 := createValidator(t, f, 2, 200)
		_, consAddr3 := createValidator(t, f, 3, 300)

		resp, err := f.poaKeeper.TotalPower(f.ctx, &types.QueryTotalPowerRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(600), resp.TotalPower) // 100 + 200 + 300

		// Verify total power equals sum of individual validator queries
		var sumOfIndividualPowers int64
		for _, consAddr := range []string{consAddr1.String(), consAddr2.String(), consAddr3.String()} {
			valResp, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
				Address: consAddr,
			})
			require.NoError(t, err)
			require.NotNil(t, valResp)
			sumOfIndividualPowers += valResp.Validator.Power
		}
		require.Equal(t, resp.TotalPower, sumOfIndividualPowers,
			"Total power should equal sum of individual validator powers")
	})

	t.Run("updates correctly when validator power changes", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		_, consAddr1 := createValidator(t, f, 1, 100)
		createValidator(t, f, 2, 200)

		// Check initial total
		resp, err := f.poaKeeper.TotalPower(f.ctx, &types.QueryTotalPowerRequest{})
		require.NoError(t, err)
		require.Equal(t, int64(300), resp.TotalPower)

		// Update first validator's power
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr1, 150)
		require.NoError(t, err)

		// Check updated total
		resp, err = f.poaKeeper.TotalPower(f.ctx, &types.QueryTotalPowerRequest{})
		require.NoError(t, err)
		require.Equal(t, int64(350), resp.TotalPower) // 150 + 200
	})
}

func TestWithdrawableFeesQuery(t *testing.T) {
	t.Run("returns zero fees for validator with no fees", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)

		// Query fees - should be zero
		resp, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.True(t, resp.Fees.Fees.IsZero())
	})

	t.Run("calculates pending fees with lazy distribution", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		opAddr1, _ := createValidator(t, f, 1, 100) // 25% power
		opAddr2, _ := createValidator(t, f, 2, 300) // 75% power

		// Add fees to fee collector
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Query validator 1 - should show 25% (250 stake) as pending
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr1,
		})
		require.NoError(t, err)
		require.Equal(t, "250.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())

		// Query validator 2 - should show 75% (750 stake) as pending
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr2,
		})
		require.NoError(t, err)
		require.Equal(t, "750.000000000000000000", resp2.Fees.Fees.AmountOf("stake").String())
	})

	t.Run("shows allocated plus pending fees after checkpoint", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)

		// Add initial fees
		fees1 := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees1)
		require.NoError(t, err)

		// Checkpoint to allocate fees
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Query should show 100 allocated
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.Equal(t, "100.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())

		// Add more fees
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Query should show 100 (allocated) + 50 (pending) = 150 total
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.Equal(t, "150.000000000000000000", resp2.Fees.Fees.AmountOf("stake").String())
	})

	t.Run("handles multiple denominations correctly", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		opAddr1, _ := createValidator(t, f, 1, 100) // 25% power
		opAddr2, _ := createValidator(t, f, 2, 300) // 75% power

		// Add multiple denominations
		fees := sdk.NewCoins(
			sdk.NewInt64Coin("stake", 1000),
			sdk.NewInt64Coin("atom", 400),
		)
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Query validator 1 - should show 25% of each denom
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr1,
		})
		require.NoError(t, err)
		require.Equal(t, "250.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())
		require.Equal(t, "100.000000000000000000", resp1.Fees.Fees.AmountOf("atom").String())

		// Query validator 2 - should show 75% of each denom
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr2,
		})
		require.NoError(t, err)
		require.Equal(t, "750.000000000000000000", resp2.Fees.Fees.AmountOf("stake").String())
		require.Equal(t, "300.000000000000000000", resp2.Fees.Fees.AmountOf("atom").String())
	})

	t.Run("returns zero for validator with zero power", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		opAddr1, _ := createValidator(t, f, 1, 100)
		opAddr2, _ := createValidator(t, f, 2, 0) // Zero power

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Validator with zero power should show zero fees
		resp, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr2,
		})
		require.NoError(t, err)
		require.True(t, resp.Fees.Fees.IsZero())

		// Validator with power should get all fees
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr1,
		})
		require.NoError(t, err)
		require.Equal(t, "1000.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())
	})

	t.Run("returns error for non-existent validator", func(t *testing.T) {
		f := setupTest(t)

		// Query for non-existent validator
		_, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: "cosmos1nonexistent",
		})
		require.Error(t, err)
	})

	t.Run("correctly calculates with partial allocation", func(t *testing.T) {
		f := setupTest(t)

		// Create validators
		opAddr1, _ := createValidator(t, f, 1, 100)
		opAddr2, _ := createValidator(t, f, 2, 100)

		// Add initial fees
		fees1 := sdk.NewCoins(sdk.NewInt64Coin("stake", 200))
		err := f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees1)
		require.NoError(t, err)

		// Checkpoint to allocate (each gets 100)
		err = f.poaKeeper.CheckpointAllValidators(f.ctx)
		require.NoError(t, err)

		// Change power distribution
		err = f.poaKeeper.SetValidatorPower(f.ctx, sdk.ConsAddress("cons1"), 300)
		require.NoError(t, err)

		// Add more fees
		fees2 := sdk.NewCoins(sdk.NewInt64Coin("stake", 400))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees2)
		require.NoError(t, err)

		// Validator 1 now has 300/(300+100) = 75% of pending
		// Should show: 100 (allocated) + 300 (75% of 400) = 400 total
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr1,
		})
		require.NoError(t, err)
		require.Equal(t, "400.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())

		// Validator 2: 100 (allocated) + 100 (25% of 400) = 200 total
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr2,
		})
		require.NoError(t, err)
		require.Equal(t, "200.000000000000000000", resp2.Fees.Fees.AmountOf("stake").String())
	})

	t.Run("shows correct fees after withdrawal", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		opAddr, _ := createValidator(t, f, 1, 100)
		opAddrSdk, err := sdk.AccAddressFromBech32(opAddr)
		require.NoError(t, err)

		// Add fees
		fees := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
		err = f.bankKeeper.MintCoins(f.ctx, authtypes.FeeCollectorName, fees)
		require.NoError(t, err)

		// Query before withdrawal
		resp1, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.Equal(t, "100.000000000000000000", resp1.Fees.Fees.AmountOf("stake").String())

		// Withdraw fees
		coins, err := f.poaKeeper.WithdrawValidatorFees(f.ctx, opAddrSdk)
		require.NoError(t, err)
		require.Equal(t, sdk.NewCoins(sdk.NewInt64Coin("stake", 100)), coins)

		// Query after withdrawal - should be zero
		resp2, err := f.poaKeeper.WithdrawableFees(f.ctx, &types.QueryWithdrawableFeesRequest{
			OperatorAddress: opAddr,
		})
		require.NoError(t, err)
		require.True(t, resp2.Fees.Fees.IsZero())
	})
}

func TestParamsQuery(t *testing.T) {
	t.Run("returns params after initialization", func(t *testing.T) {
		f := setupTest(t)

		// Initialize params first
		err := f.poaKeeper.UpdateParams(f.ctx, types.Params{Admin: sdk.AccAddress("admin").String()})
		require.NoError(t, err)

		resp, err := f.poaKeeper.Params(f.ctx, &types.QueryParamsRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Params)
	})

	t.Run("returns updated params after modification", func(t *testing.T) {
		f := setupTest(t)

		// Update params
		newParams := types.Params{
			Admin: sdk.AccAddress("admin").String(),
		}
		err := f.poaKeeper.UpdateParams(f.ctx, newParams)
		require.NoError(t, err)

		// Query params
		resp, err := f.poaKeeper.Params(f.ctx, &types.QueryParamsRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, newParams, resp.Params)
	})
}

func TestValidatorQuery(t *testing.T) {
	t.Run("returns validator by consensus address", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		opAddr, consAddr := createValidator(t, f, 1, 100)

		// Query validator by consensus address
		resp, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: consAddr.String(),
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(100), resp.Validator.Power)
		require.Equal(t, opAddr, resp.Validator.Metadata.OperatorAddress)
		require.Equal(t, "validator-1", resp.Validator.Metadata.Moniker)
	})

	t.Run("returns validator by operator address", func(t *testing.T) {
		f := setupTest(t)

		// Create validator
		opAddr, consAddr := createValidator(t, f, 1, 100)

		// Query validator by operator address
		resp, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: opAddr,
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.Equal(t, int64(100), resp.Validator.Power)
		require.Equal(t, opAddr, resp.Validator.Metadata.OperatorAddress)
		require.Equal(t, "validator-1", resp.Validator.Metadata.Moniker)

		// Verify same result as querying by consensus address
		respByCons, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: consAddr.String(),
		})
		require.NoError(t, err)
		require.Equal(t, resp.Validator, respByCons.Validator)
	})

	t.Run("returns error for non-existent validator", func(t *testing.T) {
		f := setupTest(t)

		// Create a valid consensus address that doesn't exist
		nonExistentConsAddr := sdk.ConsAddress("nonexistent")

		// Query for non-existent validator
		_, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: nonExistentConsAddr.String(),
		})
		require.Error(t, err)
	})

	t.Run("returns error for invalid address", func(t *testing.T) {
		f := setupTest(t)

		// Query with invalid address (neither consensus nor operator address)
		_, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: "invalid-address",
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "address must be either a valid consensus address or operator address")
	})

	t.Run("returns correct validator data after power update", func(t *testing.T) {
		f := setupTest(t)

		// Create validator with initial power
		opAddr, consAddr := createValidator(t, f, 1, 100)

		// Update power
		err := f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 200)
		require.NoError(t, err)

		// Query should return updated power (query by consensus address)
		resp, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: consAddr.String(),
		})
		require.NoError(t, err)
		require.Equal(t, int64(200), resp.Validator.Power)
		require.Equal(t, opAddr, resp.Validator.Metadata.OperatorAddress)

		// Query should also return updated power (query by operator address)
		respByOp, err := f.poaKeeper.Validator(f.ctx, &types.QueryValidatorRequest{
			Address: opAddr,
		})
		require.NoError(t, err)
		require.Equal(t, int64(200), respByOp.Validator.Power)
	})
}
