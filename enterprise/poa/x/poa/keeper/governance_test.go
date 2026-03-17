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

	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
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

func TestPOAVoteRemovalAfterTally(t *testing.T) {
	f := setupTest(t)

	addr1, _ := createValidator(t, f, 1, 100)
	addr2, _ := createValidator(t, f, 2, 200)
	addr3, _ := createValidator(t, f, 3, 300)

	proposerAddr, err := sdk.AccAddressFromBech32(addr1)
	require.NoError(t, err)
	proposalID := createProposal(t, f, proposerAddr)

	submitVoteDirectly(t, f, proposalID, addr1, govv1.WeightedVoteOptions{{Option: govv1.OptionYes, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID, addr2, govv1.WeightedVoteOptions{{Option: govv1.OptionNo, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID, addr3, govv1.WeightedVoteOptions{{Option: govv1.OptionAbstain, Weight: "1.0"}})

	// Verify votes exist before tally
	for _, addr := range []string{addr1, addr2, addr3} {
		voterAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		_, err = f.govKeeper.Votes.Get(f.ctx, collections.Join(proposalID, voterAddr))
		require.NoError(t, err, "vote should exist before tally")
	}

	proposal, err := f.govKeeper.Proposals.Get(f.ctx, proposalID)
	require.NoError(t, err)

	tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*f.poaKeeper)
	_, _, _, err = tallyFn(f.ctx, *f.govKeeper, proposal)
	require.NoError(t, err)

	// Votes should be removed after tally
	for _, addr := range []string{addr1, addr2, addr3} {
		voterAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		_, err = f.govKeeper.Votes.Get(f.ctx, collections.Join(proposalID, voterAddr))
		require.Error(t, err, "vote should be removed after tally")
		require.ErrorIs(t, err, collections.ErrNotFound)
	}
}

func TestPOAMultipleProposalsVoteRemoval(t *testing.T) {
	f := setupTest(t)

	addr1, _ := createValidator(t, f, 1, 100)
	addr2, _ := createValidator(t, f, 2, 200)

	proposerAddr, err := sdk.AccAddressFromBech32(addr1)
	require.NoError(t, err)

	proposalID1 := createProposal(t, f, proposerAddr)
	proposalID2 := createProposal(t, f, proposerAddr)

	// Vote on both proposals
	submitVoteDirectly(t, f, proposalID1, addr1, govv1.WeightedVoteOptions{{Option: govv1.OptionYes, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID1, addr2, govv1.WeightedVoteOptions{{Option: govv1.OptionNo, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID2, addr1, govv1.WeightedVoteOptions{{Option: govv1.OptionNo, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID2, addr2, govv1.WeightedVoteOptions{{Option: govv1.OptionYes, Weight: "1.0"}})

	// Only tally proposal1
	proposal1, err := f.govKeeper.Proposals.Get(f.ctx, proposalID1)
	require.NoError(t, err)

	tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*f.poaKeeper)
	_, _, _, err = tallyFn(f.ctx, *f.govKeeper, proposal1)
	require.NoError(t, err)

	// Proposal1 votes should be removed
	for _, addr := range []string{addr1, addr2} {
		voterAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		_, err = f.govKeeper.Votes.Get(f.ctx, collections.Join(proposalID1, voterAddr))
		require.Error(t, err)
		require.ErrorIs(t, err, collections.ErrNotFound)
	}

	// Proposal2 votes should still exist
	for _, addr := range []string{addr1, addr2} {
		voterAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		_, err = f.govKeeper.Votes.Get(f.ctx, collections.Join(proposalID2, voterAddr))
		require.NoError(t, err, "proposal2 votes should still exist")
	}
}

func TestPOADepoweredValidatorVoteRemoval(t *testing.T) {
	f := setupTest(t)

	addr1, consAddr1 := createValidator(t, f, 1, 100)
	addr2, _ := createValidator(t, f, 2, 200)

	proposerAddr, err := sdk.AccAddressFromBech32(addr1)
	require.NoError(t, err)
	proposalID := createProposal(t, f, proposerAddr)

	// Both validators vote
	submitVoteDirectly(t, f, proposalID, addr1, govv1.WeightedVoteOptions{{Option: govv1.OptionYes, Weight: "1.0"}})
	submitVoteDirectly(t, f, proposalID, addr2, govv1.WeightedVoteOptions{{Option: govv1.OptionNo, Weight: "1.0"}})

	// De-power validator1 before tally
	err = f.poaKeeper.SetValidatorPower(f.ctx, consAddr1, 0)
	require.NoError(t, err)

	proposal, err := f.govKeeper.Proposals.Get(f.ctx, proposalID)
	require.NoError(t, err)

	tallyFn := NewPOACalculateVoteResultsAndVotingPowerFn(*f.poaKeeper)
	totalVoterPower, totalValPower, results, err := tallyFn(f.ctx, *f.govKeeper, proposal)
	require.NoError(t, err)

	// Only validator2's vote should be tallied
	require.Equal(t, math.LegacyNewDec(200), totalVoterPower)
	require.Equal(t, math.NewInt(200), totalValPower)
	require.Equal(t, math.LegacyZeroDec(), results[govv1.OptionYes])
	require.Equal(t, math.LegacyNewDec(200), results[govv1.OptionNo])

	// Both votes should be removed, including the de-powered validator's vote
	for _, addr := range []string{addr1, addr2} {
		voterAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		_, err = f.govKeeper.Votes.Get(f.ctx, collections.Join(proposalID, voterAddr))
		require.Error(t, err, "vote should be removed after tally even for de-powered validator")
		require.ErrorIs(t, err, collections.ErrNotFound)
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
