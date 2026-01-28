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
	"context"
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/enterprise/poa/x/poa/types"
)

// NewPOACalculateVoteResultsAndVotingPowerFn returns a custom vote tallying function for POA governance.
// Unlike standard governance which uses bonded stake, this function uses validator power as voting weight.
// Only active validators (power > 0) have their votes counted.
// The function returns:
//   - totalVoterPower: sum of power from validators who voted
//   - totalValPower: total power of all active validators
//   - results: vote counts per option, weighted by validator power
func NewPOACalculateVoteResultsAndVotingPowerFn(keeper Keeper) govkeeper.CalculateVoteResultsAndVotingPowerFn {
	return func(ctx context.Context, k govkeeper.Keeper, proposal govv1.Proposal) (totalVoterPower math.LegacyDec, totalValPower math.Int, results map[govv1.VoteOption]math.LegacyDec, err error) {
		sdkCtx := sdk.UnwrapSDKContext(ctx)

		// Initialize vote results
		results = make(map[govv1.VoteOption]math.LegacyDec)
		results[govv1.OptionYes] = math.LegacyZeroDec()
		results[govv1.OptionAbstain] = math.LegacyZeroDec()
		results[govv1.OptionNo] = math.LegacyZeroDec()
		results[govv1.OptionNoWithVeto] = math.LegacyZeroDec()

		totalVoterPower = math.LegacyZeroDec()

		// Get total validator power from storage
		totalPower, err := keeper.GetTotalPower(sdkCtx)
		if err != nil {
			return math.LegacyZeroDec(), math.ZeroInt(), nil, err
		}
		totalValPower = math.NewInt(totalPower)

		// Collect all validators and their voting power
		validators := make(map[string]int64) // operator address -> power

		// Iterate through POA validators in descending power order (highest power first)
		// Stop when we reach power = 0
		ranger := new(collections.Range[collections.Pair[int64, string]]).Descending()
		err = keeper.validators.Walk(sdkCtx, ranger, func(key collections.Pair[int64, string], validator types.Validator) (stop bool, err error) {
			power := key.K1()

			// Stop iteration when we reach validators with power 0
			if power == 0 {
				return true, nil
			}

			validators[validator.Metadata.OperatorAddress] = power

			return false, nil
		})
		if err != nil {
			return math.LegacyZeroDec(), math.ZeroInt(), nil, err
		}

		// Iterate through all votes for this proposal
		rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
		err = k.Votes.Walk(sdkCtx, rng, func(key collections.Pair[uint64, sdk.AccAddress], vote govv1.Vote) (bool, error) {
			// Check if the voter is a POA validator
			power, isValidator := validators[vote.Voter]
			if !isValidator {
				// Skip non-validator votes in POA
				// Note that this should never happen if POA governance is set up properly
				return false, nil
			}

			// Calculate voting power for this validator
			votingPower := math.LegacyNewDec(power)

			// Tally the vote with weighted options
			for _, option := range vote.Options {
				weight, _ := math.LegacyNewDecFromStr(option.Weight)
				optionPower := votingPower.Mul(weight)
				results[option.Option] = results[option.Option].Add(optionPower)
			}

			// Add to total voter power
			totalVoterPower = totalVoterPower.Add(votingPower)

			return false, nil
		})
		if err != nil {
			return math.LegacyZeroDec(), math.ZeroInt(), nil, err
		}

		return totalVoterPower, totalValPower, results, nil
	}
}

// validateAuthorizedValidator checks if an address is allowed to participate in POA governance.
// Only registered POA validators with positive voting power are authorized.
func (k *Keeper) validateAuthorizedValidator(ctx sdk.Context, voterAddr sdk.AccAddress) error {
	voterAddrStr := voterAddr.String()

	// Use the secondary index to find the validator by operator address
	validator, err := k.GetValidatorByOperatorAddress(ctx, voterAddr)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return fmt.Errorf("voter %s is not an active POA validator", voterAddrStr)
		}
		return fmt.Errorf("error validating voter: %w", err)
	}

	// Check if the validator has voting power
	if validator.Power <= 0 {
		return fmt.Errorf("voter %s is not an active POA validator", voterAddrStr)
	}

	return nil
}
