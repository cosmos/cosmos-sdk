package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	v1 "cosmossdk.io/x/gov/types/v1"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Tally iterates over the votes and updates the tally of a proposal based on the voting power of the voters
func (keeper Keeper) Tally(ctx context.Context, proposal v1.Proposal) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	validators, err := keeper.getCurrentValidators(ctx)
	if err != nil {
		return false, false, v1.TallyResult{}, err
	}

	totalVoterPower, results, err := keeper.calculateVoteResultsAndVotingPower(ctx, proposal.Id, validators)
	if err != nil {
		return false, false, v1.TallyResult{}, err
	}

	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return false, false, v1.TallyResult{}, err
	}
	tallyResults = v1.NewTallyResultFromMap(results)

	// If there is no staked coins, the proposal fails
	totalBonded, err := keeper.sk.TotalBondedTokens(ctx)
	if err != nil {
		return false, false, v1.TallyResult{}, err
	}

	if totalBonded.IsZero() {
		return false, false, tallyResults, nil
	}

	// If there are more spam votes than the sum of all other options, proposal fails
	// A proposal with no votes should not be considered spam
	if !totalVoterPower.Equal(math.LegacyZeroDec()) &&
		results[v1.OptionSpam].GTE(results[v1.OptionOne].Add(results[v1.OptionTwo].Add(results[v1.OptionThree].Add(results[v1.OptionFour])))) {
		return false, true, tallyResults, nil
	}

	switch proposal.ProposalType {
	case v1.ProposalType_PROPOSAL_TYPE_OPTIMISTIC:
		return keeper.tallyOptimistic(totalVoterPower, totalBonded, results, params)
	case v1.ProposalType_PROPOSAL_TYPE_EXPEDITED:
		return keeper.tallyExpedited(totalVoterPower, totalBonded, results, params)
	case v1.ProposalType_PROPOSAL_TYPE_MULTIPLE_CHOICE:
		return keeper.tallyMultipleChoice(totalVoterPower, totalBonded, results, params) // TODO(@julienrbrt): implement in follow up
	default:
		return keeper.tallyStandard(totalVoterPower, totalBonded, results, params)
	}
}

// tallyStandard tallies the votes of a standard proposal
func (keeper Keeper) tallyStandard(totalVoterPower math.LegacyDec, totalBonded math.Int, results map[v1.VoteOption]math.LegacyDec, params v1.Params) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	tallyResults = v1.NewTallyResultFromMap(results)

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVoterPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := math.LegacyNewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults, nil
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVoterPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return false, false, tallyResults, nil
	}

	// If more than 1/3 of voters veto, proposal fails
	vetoThreshold, _ := math.LegacyNewDecFromStr(params.VetoThreshold)
	if results[v1.OptionNoWithVeto].Quo(totalVoterPower).GT(vetoThreshold) {
		return false, params.BurnVoteVeto, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	threshold, _ := math.LegacyNewDecFromStr(params.GetThreshold())

	if results[v1.OptionYes].Quo(totalVoterPower.Sub(results[v1.OptionAbstain])).GT(threshold) {
		return true, false, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults, nil
}

// tallyExpedited tallies the votes of an expedited proposal
func (keeper Keeper) tallyExpedited(totalVoterPower math.LegacyDec, totalBonded math.Int, results map[v1.VoteOption]math.LegacyDec, params v1.Params) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	tallyResults = v1.NewTallyResultFromMap(results)

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVoterPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := math.LegacyNewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults, nil
	}

	// If no one votes (everyone abstains), proposal fails
	if totalVoterPower.Sub(results[v1.OptionAbstain]).Equal(math.LegacyZeroDec()) {
		return false, false, tallyResults, nil
	}

	// If more than 1/3 of voters veto, proposal fails
	vetoThreshold, _ := math.LegacyNewDecFromStr(params.VetoThreshold)
	if results[v1.OptionNoWithVeto].Quo(totalVoterPower).GT(vetoThreshold) {
		return false, params.BurnVoteVeto, tallyResults, nil
	}

	// If more than 2/3 of non-abstaining voters vote Yes, proposal passes
	threshold, _ := math.LegacyNewDecFromStr(params.GetExpeditedThreshold())

	if results[v1.OptionYes].Quo(totalVoterPower.Sub(results[v1.OptionAbstain])).GT(threshold) {
		return true, false, tallyResults, nil
	}

	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, false, tallyResults, nil
}

// tallyOptimistic tallies the votes of an optimistic proposal
func (keeper Keeper) tallyOptimistic(totalVoterPower math.LegacyDec, totalBonded math.Int, results map[v1.VoteOption]math.LegacyDec, params v1.Params) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	tallyResults = v1.NewTallyResultFromMap(results)
	optimisticNoThreshold, _ := math.LegacyNewDecFromStr(params.OptimisticRejectedThreshold)

	// If proposal has no votes, proposal passes
	if totalVoterPower.Equal(math.LegacyZeroDec()) {
		return true, false, tallyResults, nil
	}

	// If the threshold of no is reached, proposal fails
	if results[v1.OptionNo].Quo(totalBonded.ToLegacyDec()).GT(optimisticNoThreshold) {
		return false, false, tallyResults, nil
	}

	return true, false, tallyResults, nil
}

// tallyMultipleChoice tallies the votes of a multiple choice proposal
func (keeper Keeper) tallyMultipleChoice(totalVoterPower math.LegacyDec, totalBonded math.Int, results map[v1.VoteOption]math.LegacyDec, params v1.Params) (passes, burnDeposits bool, tallyResults v1.TallyResult, err error) {
	tallyResults = v1.NewTallyResultFromMap(results)

	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVoterPower.Quo(math.LegacyNewDecFromInt(totalBonded))
	quorum, _ := math.LegacyNewDecFromStr(params.Quorum)
	if percentVoting.LT(quorum) {
		return false, params.BurnVoteQuorum, tallyResults, nil
	}

	return true, false, tallyResults, nil
}

// getCurrentValidators fetches all the bonded validators, insert them into currValidators
func (keeper Keeper) getCurrentValidators(ctx context.Context) (map[string]v1.ValidatorGovInfo, error) {
	currValidators := make(map[string]v1.ValidatorGovInfo)
	if err := keeper.sk.IterateBondedValidatorsByPower(ctx, func(index int64, validator sdk.ValidatorI) (stop bool) {
		valBz, err := keeper.sk.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
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

// calculateVoteResultsAndVotingPower iterate over all votes, tally up the voting power of each validator
// and returns the votes results from voters
func (keeper Keeper) calculateVoteResultsAndVotingPower(
	ctx context.Context,
	proposalID uint64,
	validators map[string]v1.ValidatorGovInfo,
) (math.LegacyDec, map[v1.VoteOption]math.LegacyDec, error) {
	totalVP := math.LegacyZeroDec()
	results := createEmptyResults()

	// iterate over all votes, tally up the voting power of each validator
	rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposalID)
	if err := keeper.Votes.Walk(ctx, rng, func(key collections.Pair[uint64, sdk.AccAddress], vote v1.Vote) (bool, error) {
		// if validator, just record it in the map
		voter, err := keeper.authKeeper.AddressCodec().StringToBytes(vote.Voter)
		if err != nil {
			return false, err
		}

		valAddrStr, err := keeper.sk.ValidatorAddressCodec().BytesToString(voter)
		if err != nil {
			return false, err
		}

		if val, ok := validators[valAddrStr]; ok {
			val.Vote = vote.Options
			validators[valAddrStr] = val
		}

		// iterate over all delegations from voter, deduct from any delegated-to validators
		err = keeper.sk.IterateDelegations(ctx, voter, func(index int64, delegation sdk.DelegationI) (stop bool) {
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

				totalVP = totalVP.Add(votingPower)
			}

			return false
		})
		if err != nil {
			return false, err
		}

		return false, keeper.Votes.Remove(ctx, collections.Join(vote.ProposalId, sdk.AccAddress(voter)))
	}); err != nil {
		return math.LegacyDec{}, nil, err
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
		totalVP = totalVP.Add(votingPower)
	}

	return totalVP, results, nil
}

func createEmptyResults() map[v1.VoteOption]math.LegacyDec {
	results := make(map[v1.VoteOption]math.LegacyDec)
	results[v1.OptionYes] = math.LegacyZeroDec()
	results[v1.OptionAbstain] = math.LegacyZeroDec()
	results[v1.OptionNo] = math.LegacyZeroDec()
	results[v1.OptionNoWithVeto] = math.LegacyZeroDec()
	results[v1.OptionSpam] = math.LegacyZeroDec()

	return results
}
