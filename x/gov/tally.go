package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// validatorGovInfo used for tallying
type validatorGovInfo struct {
	Address             sdk.ValAddress // address of the validator operator
	BondedTokens        sdk.Int        // Power of a Validator
	DelegatorShares     sdk.Dec        // Total outstanding delegator shares
	DelegatorDeductions sdk.Dec        // Delegator deductions from validator's delegators voting independently
	Vote                VoteOption     // Vote of the validator
}

func newValidatorGovInfo(address sdk.ValAddress, bondedTokens sdk.Int, delegatorShares,
	delegatorDeductions sdk.Dec, vote VoteOption) validatorGovInfo {

	return validatorGovInfo{
		Address:             address,
		BondedTokens:        bondedTokens,
		DelegatorShares:     delegatorShares,
		DelegatorDeductions: delegatorDeductions,
		Vote:                vote,
	}
}

// TODO: Break into several smaller functions for clarity
func tally(ctx sdk.Context, keeper Keeper, proposal Proposal) (passes bool, tallyResults TallyResult) {
	results := make(map[VoteOption]sdk.Dec)
	results[OptionYes] = sdk.ZeroDec()
	results[OptionAbstain] = sdk.ZeroDec()
	results[OptionNo] = sdk.ZeroDec()
	results[OptionNoWithVeto] = sdk.ZeroDec()

	totalVotingPower := sdk.ZeroDec()
	currValidators := make(map[string]validatorGovInfo)

	// fetch all the bonded validators, insert them into currValidators
	keeper.vs.IterateBondedValidatorsByPower(ctx, func(index int64, validator sdk.Validator) (stop bool) {
		currValidators[validator.GetOperator().String()] = newValidatorGovInfo(
			validator.GetOperator(),
			validator.GetBondedTokens(),
			validator.GetDelegatorShares(),
			sdk.ZeroDec(),
			OptionEmpty,
		)
		return false
	})

	// iterate over all the votes
	votesIterator := keeper.GetVotes(ctx, proposal.GetProposalID())
	defer votesIterator.Close()
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(votesIterator.Value(), vote)

		// if validator, just record it in the map
		// if delegator tally voting power
		valAddrStr := sdk.ValAddress(vote.Voter).String()
		if val, ok := currValidators[valAddrStr]; ok {
			val.Vote = vote.Option
			currValidators[valAddrStr] = val
		} else {
			// iterate over all delegations from voter, deduct from any delegated-to validators
			keeper.ds.IterateDelegations(ctx, vote.Voter, func(index int64, delegation sdk.Delegation) (stop bool) {
				valAddrStr := delegation.GetValidatorAddr().String()

				if val, ok := currValidators[valAddrStr]; ok {
					val.DelegatorDeductions = val.DelegatorDeductions.Add(delegation.GetShares())
					currValidators[valAddrStr] = val

					delegatorShare := delegation.GetShares().Quo(val.DelegatorShares)
					votingPower := delegatorShare.MulInt(val.BondedTokens)

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

		sharesAfterDeductions := val.DelegatorShares.Sub(val.DelegatorDeductions)
		fractionAfterDeductions := sharesAfterDeductions.Quo(val.DelegatorShares)
		votingPower := fractionAfterDeductions.MulInt(val.BondedTokens)

		results[val.Vote] = results[val.Vote].Add(votingPower)
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	tallyParams := keeper.GetTallyParams(ctx)
	tallyResults = NewTallyResultFromMap(results)

	// TODO: Upgrade the spec to cover all of these cases & remove pseudocode.
	// If there is no staked coins, the proposal fails
	if keeper.vs.TotalBondedTokens(ctx).IsZero() {
		return false, tallyResults
	}
	// If there is not enough quorum of votes, the proposal fails
	percentVoting := totalVotingPower.Quo(sdk.NewDecFromInt(keeper.vs.TotalBondedTokens(ctx)))
	if percentVoting.LT(tallyParams.Quorum) {
		return false, tallyResults
	}
	// If no one votes (everyone abstains), proposal fails
	if totalVotingPower.Sub(results[OptionAbstain]).Equal(sdk.ZeroDec()) {
		return false, tallyResults
	}
	// If more than 1/3 of voters veto, proposal fails
	if results[OptionNoWithVeto].Quo(totalVotingPower).GT(tallyParams.Veto) {
		return false, tallyResults
	}
	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[OptionYes].Quo(totalVotingPower.Sub(results[OptionAbstain])).GT(tallyParams.Threshold) {
		return true, tallyResults
	}
	// If more than 1/2 of non-abstaining voters vote No, proposal fails

	return false, tallyResults
}
