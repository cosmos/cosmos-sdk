package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// validatorGovInfo used for tallying
type validatorGovInfo struct {
	Address         sdk.ValAddress // address of the validator operator
	Power           sdk.Dec        // Power of a Validator
	DelegatorShares sdk.Dec        // Total outstanding delegator shares
	Minus           sdk.Dec        // Minus of validator, used to compute validator's voting power
	Vote            VoteOption     // Vote of the validator
}

func tally(ctx sdk.Context, keeper Keeper, proposal Proposal) (passes bool, tallyResults TallyResult) {
	results := make(map[VoteOption]sdk.Dec)
	results[OptionYes] = sdk.ZeroDec()
	results[OptionAbstain] = sdk.ZeroDec()
	results[OptionNo] = sdk.ZeroDec()
	results[OptionNoWithVeto] = sdk.ZeroDec()

	totalVotingPower := sdk.ZeroDec()
	currValidators := make(map[string]validatorGovInfo)

	keeper.vs.IterateValidatorsBonded(ctx, func(index int64, validator sdk.Validator) (stop bool) {
		currValidators[validator.GetOperator().String()] = validatorGovInfo{
			Address:         validator.GetOperator(),
			Power:           validator.GetPower(),
			DelegatorShares: validator.GetDelegatorShares(),
			Minus:           sdk.ZeroDec(),
			Vote:            OptionEmpty,
		}
		return false
	})

	// iterate over all the votes
	votesIterator := keeper.GetVotes(ctx, proposal.GetProposalID())
	defer votesIterator.Close()
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)

		// if validator, just record it in the map
		// if delegator tally voting power
		valAddrStr := sdk.ValAddress(vote.Voter).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Option
			currValidators[valAddrStr] = val
		} else {

			keeper.ds.IterateDelegations(ctx, vote.Voter, func(index int64, delegation sdk.Delegation) (stop bool) {
				valAddrStr := delegation.GetValidatorAddr().String()

				if val, ok := currValidators[valAddrStr]; ok {
					val.Minus = val.Minus.Add(delegation.GetShares())
					currValidators[valAddrStr] = val

					delegatorShare := delegation.GetShares().Quo(val.DelegatorShares)
					votingPower := val.Power.Mul(delegatorShare)

					results[vote.Option] = results[vote.Option].Add(votingPower)
					totalVotingPower = totalVotingPower.Add(votingPower)
				}

				return false
			})
		}

		keeper.deleteVote(ctx, vote.ProposalID, vote.Voter)
	}

	// iterate over the validators again to tally their voting power
	for _, val := range currValidators {
		if val.Vote == OptionEmpty {
			continue
		}

		sharesAfterMinus := val.DelegatorShares.Sub(val.Minus)
		percentAfterMinus := sharesAfterMinus.Quo(val.DelegatorShares)
		votingPower := val.Power.Mul(percentAfterMinus)

		results[val.Vote] = results[val.Vote].Add(votingPower)
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	tallyingProcedure := keeper.GetTallyingProcedure(ctx)

	tallyResults = TallyResult{
		Yes:        results[OptionYes],
		Abstain:    results[OptionAbstain],
		No:         results[OptionNo],
		NoWithVeto: results[OptionNoWithVeto],
	}

	// If no one votes, proposal fails
	if totalVotingPower.Sub(results[OptionAbstain]).Equal(sdk.ZeroDec()) {
		return false, tallyResults
	}
	// If more than 1/3 of voters veto, proposal fails
	if results[OptionNoWithVeto].Quo(totalVotingPower).GT(tallyingProcedure.Veto) {
		return false, tallyResults
	}
	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[OptionYes].Quo(totalVotingPower.Sub(results[OptionAbstain])).GT(tallyingProcedure.Threshold) {
		return true, tallyResults
	}
	// If more than 1/2 of non-abstaining voters vote No, proposal fails

	return false, tallyResults
}
