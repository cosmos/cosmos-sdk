package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// CalculateVoteResultsAndVotingPowerFn is a function signature for calculating vote results and voting power
// It can be overridden to customize the voting power calculation for proposals
// It gets the proposal tallied and the validators governance infos (bonded tokens, voting power, etc.)
// It must return the total voting power and the results of the vote
type CalculateVoteResultsAndVotingPowerFn func(
	ctx context.Context,
	k Keeper,
	proposal v1.Proposal,
	validators map[string]v1.ValidatorGovInfo,
) (totalVoterPower math.LegacyDec, results map[v1.VoteOption]math.LegacyDec, err error)

func defaultCalculateVoteResultsAndVotingPower(
	ctx context.Context,
	k Keeper,
	proposal v1.Proposal,
	validators map[string]v1.ValidatorGovInfo,
) (totalVoterPower math.LegacyDec, results map[v1.VoteOption]math.LegacyDec, err error) {
	totalVotingPower := math.LegacyZeroDec()

	results = make(map[v1.VoteOption]math.LegacyDec)
	results[v1.OptionYes] = math.LegacyZeroDec()
	results[v1.OptionAbstain] = math.LegacyZeroDec()
	results[v1.OptionNo] = math.LegacyZeroDec()
	results[v1.OptionNoWithVeto] = math.LegacyZeroDec()

	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
	votesToRemove := []collections.Pair[uint64, sdk.AccAddress]{}
	err = k.Votes.Walk(ctx, rng, func(key collections.Pair[uint64, sdk.AccAddress], vote v1.Vote) (bool, error) {
		// if validator, just record it in the map
		voter, err := k.authKeeper.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			return false, err
		}

		valAddrStr, err := k.sk.ValidatorAddressCodec().BytesToString(voter)
		if err != nil {
			return false, err
		}
		if val, ok := validators[valAddrStr]; ok {
			val.Vote = vote.Options
			validators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		err = k.sk.IterateDelegations(ctx, voter, func(index int64, delegation stakingtypes.DelegationI) (stop bool) {
			valAddrStr := delegation.GetValidatorAddr()

			if val, ok := validators[valAddrStr]; ok {
				// There is no need to handle the special case that validator address equal to voter address.
				// Because voter's voting power will tally again even if there will be deduction of voter's voting power from validator.
				val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
				validators[valAddrStr] = val

				// delegation shares * bonded / total shares
				votingPower := delegation.GetShares().MulInt(val.BondedTokens).Quo(val.DelegatorShares)

				for _, option := range vote.Options {
					weight, _ := math.LegacyNewDecFromStr(option.Weight)
					subPower := votingPower.Mul(weight)
					results[option.Option] = results[option.Option].Add(subPower)
				}
				totalVotingPower = totalVotingPower.Add(votingPower)
			}

			return false
		})
		if err != nil {
			return false, err
		}

		votesToRemove = append(votesToRemove, key)
		return false, nil
	})
	if err != nil {
		return math.LegacyZeroDec(), nil, fmt.Errorf("error while iterating delegations: %w", err)
	}

	// remove all votes from store
	for _, key := range votesToRemove {
		if err := k.Votes.Remove(ctx, key); err != nil {
			return math.LegacyDec{}, nil, fmt.Errorf("error while removing vote (%d/%s): %w", key.K1(), key.K2(), err)
		}
	}

	// iterate over the validators again to tally their voting power
	for _, val := range validators {
		if len(val.Vote) == 0 {
			continue
		}

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		votingPower := sharesAfterDeductions.MulInt(val.BondedTokens).Quo(val.DelegatorShares)

		for _, option := range val.Vote {
			weight, _ := math.LegacyNewDecFromStr(option.Weight)
			subPower := votingPower.Mul(weight)
			results[option.Option] = results[option.Option].Add(subPower)
		}
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	return totalVotingPower, results, nil
}

// getCurrentValidators fetches all the bonded validators, insert them into currValidators
func (k Keeper) getCurrentValidators(ctx context.Context) (map[string]v1.ValidatorGovInfo, error) {
	currValidators := make(map[string]v1.ValidatorGovInfo)
	if err := k.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {
		valBz, err := k.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
		if err != nil {
			return false
		}
		currValidators[validator.GetOperator()] = v1.NewValidatorGovInfo(
			valBz,
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			math.LegacyZeroDec(),
			v1.WeightedVoteOptions{},
		)

		return false
	}); err != nil {
		return nil, err
	}

	return currValidators, nil
}

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the
// voters
func (k Keeper) Tally(ctx context.Context, proposal v1.Proposal) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	currValidators, err := k.getCurrentValidators(ctx)
	if err != nil {
		return false, false, tallyResults, fmt.Errorf("error while getting current validators: %w", err)
	}

	tallyFn := k.calculateVoteResultsAndVotingPowerFn
	totalVotingPower, results, err := tallyFn(ctx, k, proposal, currValidators)
	if err != nil {
		return false, false, tallyResults, fmt.Errorf("error while calculating tally results: %w", err)
	}

	tallyResults = v1.NewTallyResultFromMap(results)

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	totalBonded, err := k.sk.TotalBondedTokens(ctx)
	if err != nil {
		return false, false, tallyResults, err
	}

	if totalBonded.IsZero() {
		return false, false, tallyResults, nil
	}

	params, err := k.Params.Get(ctx)
	if err != nil {
		return false, false, tallyResults, fmt.Errorf("error while getting params: %w", err)
	}

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := math.LegacyNewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults, nil
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return false, false, tallyResults, nil
	}

	// If more than 1/3 of voters veto, proposal fails
	vetoThreshold, _ := math.LegacyNewDecFromStr(params.VetoThreshold)
	if results[v1.OptionNoWithVeto].Quo(totalVotingPower).GT(vetoThreshold) {
		return false, params.BurnVoteVeto, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	// For expedited 2/3
	var thresholdStr string
	if proposal.Expedited {
		thresholdStr = params.GetExpeditedThreshold()
	} else {
		thresholdStr = params.GetThreshold()
	}

	threshold, _ := math.LegacyNewDecFromStr(thresholdStr)

	if results[v1.OptionYes].Quo(totalVotingPower.Sub(results[v1.OptionAbstain])).GT(threshold) {
		return true, false, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults, nil
}
