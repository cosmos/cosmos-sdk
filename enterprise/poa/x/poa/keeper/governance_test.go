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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestPOACalculateVoteResultsAndVotingPower(t *testing.T) {
	tests := []struct {
		name               string
		validators         map[int]int64                     // validator index -> power
		votes              map[int]govv1.WeightedVoteOptions // validator index -> vote
		expectedResults    map[govv1.VoteOption]math.LegacyDec
		expectedVoterPower math.LegacyDec
		expectedValPower   math.Int
	}{
		{
			name: "single validator votes yes",
			validators: map[int]int64{
				1: 100,
			},
			votes: map[int]govv1.WeightedVoteOptions{
				1: {
					{Option: govv1.OptionYes, Weight: "1.0"},
				},
			},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyNewDec(100),
				govv1.OptionNo:         math.LegacyZeroDec(),
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyNewDec(100),
			expectedValPower:   math.NewInt(100),
		},
		{
			name: "multiple validators with different powers",
			validators: map[int]int64{
				1: 100,
				2: 200,
				3: 300,
			},
			votes: map[int]govv1.WeightedVoteOptions{
				1: {{Option: govv1.OptionYes, Weight: "1.0"}},
				2: {{Option: govv1.OptionNo, Weight: "1.0"}},
				3: {{Option: govv1.OptionYes, Weight: "1.0"}},
			},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyNewDec(400), // 100 + 300
				govv1.OptionNo:         math.LegacyNewDec(200),
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyNewDec(600),
			expectedValPower:   math.NewInt(600),
		},
		{
			name: "weighted vote (split vote)",
			validators: map[int]int64{
				1: 100,
			},
			votes: map[int]govv1.WeightedVoteOptions{
				1: {
					{Option: govv1.OptionYes, Weight: "0.6"},
					{Option: govv1.OptionNo, Weight: "0.4"},
				},
			},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyNewDecWithPrec(60, 0), // 100 * 0.6
				govv1.OptionNo:         math.LegacyNewDecWithPrec(40, 0), // 100 * 0.4
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyNewDec(100),
			expectedValPower:   math.NewInt(100),
		},
		{
			name: "validator doesn't vote - should not count in voter power",
			validators: map[int]int64{
				1: 100,
				2: 200,
			},
			votes: map[int]govv1.WeightedVoteOptions{
				1: {{Option: govv1.OptionYes, Weight: "1.0"}},
				// validator2 doesn't vote
			},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyNewDec(100),
				govv1.OptionNo:         math.LegacyZeroDec(),
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyNewDec(100), // only validator1 voted
			expectedValPower:   math.NewInt(300),       // total power of all validators
		},
		{
			name: "non-validator vote should be ignored",
			validators: map[int]int64{
				1: 100,
			},
			votes: map[int]govv1.WeightedVoteOptions{
				1:   {{Option: govv1.OptionYes, Weight: "1.0"}},
				999: {{Option: govv1.OptionNo, Weight: "1.0"}}, // non-validator - should be ignored
			},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyNewDec(100),
				govv1.OptionNo:         math.LegacyZeroDec(), // non-validator vote ignored
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyNewDec(100),
			expectedValPower:   math.NewInt(100),
		},
		{
			name: "no votes cast",
			validators: map[int]int64{
				1: 100,
				2: 200,
			},
			votes: map[int]govv1.WeightedVoteOptions{},
			expectedResults: map[govv1.VoteOption]math.LegacyDec{
				govv1.OptionYes:        math.LegacyZeroDec(),
				govv1.OptionNo:         math.LegacyZeroDec(),
				govv1.OptionAbstain:    math.LegacyZeroDec(),
				govv1.OptionNoWithVeto: math.LegacyZeroDec(),
			},
			expectedVoterPower: math.LegacyZeroDec(),
			expectedValPower:   math.NewInt(300),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setupTest(t)

			// Create validators in the POA keeper and track their addresses
			validatorAddrs := make(map[int]string)
			var proposerAddr sdk.AccAddress
			for idx, power := range tt.validators {
				addr, _ := createValidator(t, f, idx, power)
				validatorAddrs[idx] = addr
				// Use the first validator as proposer
				if proposerAddr == nil {
					proposerAddrBech32, err := sdk.AccAddressFromBech32(addr)
					require.NoError(t, err)
					proposerAddr = proposerAddrBech32
				}
			}

			// Create a proposal (using first validator as proposer)
			proposalID := createProposal(t, f, proposerAddr)

			// Submit votes directly to bypass hooks (for tally testing)
			for voterIdx, options := range tt.votes {
				voterAddr := validatorAddrs[voterIdx]
				if voterAddr == "" {
					// Create a non-validator voter address
					nonValidatorAddr := sdk.AccAddress(fmt.Sprintf("nonvalidator%d", voterIdx))
					voterAddr = nonValidatorAddr.String()
				}
				submitVoteDirectly(t, f, proposalID, voterAddr, options)
			}

			// Get the proposal
			proposal, err := f.govKeeper.Proposals.Get(f.ctx, proposalID)
			require.NoError(t, err)

			// Create the tally function
			tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*f.poaKeeper)

			// Execute the tally
			totalVoterPower, totalValPower, results, err := tallyFn(f.ctx, *f.govKeeper, proposal)
			require.NoError(t, err)

			// Assert results
			require.Equal(t, tt.expectedVoterPower, totalVoterPower, "total voter power mismatch")
			require.Equal(t, tt.expectedValPower, totalValPower, "total validator power mismatch")

			// Check each vote option
			for option, expectedValue := range tt.expectedResults {
				actualValue, ok := results[option]
				require.True(t, ok, "vote option %v not found in results", option)
				require.Equal(t, expectedValue, actualValue, "vote tally mismatch for option %v", option)
			}
		})
	}
}

func TestPOACalculateVoteResultsAndVotingPower_EdgeCases(t *testing.T) {
	t.Run("cannot set last validator to zero power", func(t *testing.T) {
		f := setupTest(t)

		// Create a temporary validator just to submit the proposal
		proposerAddr, _ := createValidator(t, f, 1, 100)
		proposerAddrBech32, err := sdk.AccAddressFromBech32(proposerAddr)
		require.NoError(t, err)

		proposalID := createProposal(t, f, proposerAddrBech32)
		proposal, err := f.govKeeper.Proposals.Get(f.ctx, proposalID)
		require.NoError(t, err)

		// Try to remove the last validator by setting power to 0 - should fail
		consAddr := sdk.ConsAddress("cons1")
		err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr, 0)
		require.Error(t, err)
		require.ErrorIs(t, err, types.ErrInvalidTotalPower)

		// Verify validator power was NOT changed (validation failed before changes)
		power, err := f.poaKeeper.GetValidatorPower(f.ctx, consAddr)
		require.NoError(t, err)
		require.Equal(t, int64(100), power)

		// Tally should work normally with validator still having power
		tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*f.poaKeeper)
		totalVoterPower, totalValPower, results, err := tallyFn(f.ctx, *f.govKeeper, proposal)

		require.NoError(t, err)
		require.Equal(t, math.LegacyZeroDec(), totalVoterPower) // No votes cast
		require.Equal(t, math.NewInt(100), totalValPower)       // Validator still has power
		require.Equal(t, math.LegacyZeroDec(), results[govv1.OptionYes])
		require.Equal(t, math.LegacyZeroDec(), results[govv1.OptionNo])
	})
}
