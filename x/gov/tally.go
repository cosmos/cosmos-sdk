package gov

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// validatorGovInfo used for tallying
type validatorGovInfo struct {
	ValidatorInfo stake.Validator //  Voting power of validator when proposal enters voting period
	Minus         sdk.Rat         //  Minus of validator, used to compute validator's voting power
	Vote          VoteOption      // Vote of the validator
	Power         sdk.Rat         // Power of a Validator
}

func tally(ctx sdk.Context, keeper Keeper, proposal Proposal) (passes bool, nonVoting []sdk.Address) {
	results := make(map[string]sdk.Rat)
	results[VoteOptionToString(OptionYes)] = sdk.ZeroRat()
	results[VoteOptionToString(OptionAbstain)] = sdk.ZeroRat()
	results[VoteOptionToString(OptionNo)] = sdk.ZeroRat()
	results[VoteOptionToString(OptionNoWithVeto)] = sdk.ZeroRat()

	pool := keeper.sk.GetPool(ctx)

	totalVotingPower := sdk.ZeroRat()
	currValidators := make(map[string]validatorGovInfo)
	// Gets the info on each validator and load the map
	for _, val := range keeper.sk.GetValidatorsBonded(ctx) {
		currValidators[string(val.Owner)] = validatorGovInfo{
			ValidatorInfo: val,
			Minus:         sdk.ZeroRat(),
		}
	}

	// iterate over all the votes
	votesIterator := keeper.GetVotes(ctx, proposal.GetProposalID())
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)

		// if validator, just record it in the map
		// if delegator tally voting power
		if val, ok := currValidators[string(vote.Voter)]; ok {
			val.Vote = vote.Option
			currValidators[string(vote.Voter)] = val
		} else {
			for _, delegation := range keeper.sk.GetDelegations(ctx, vote.Voter, math.MaxInt16) { // TODO: Replace with MaxValidators from Stake params
				val := currValidators[string(delegation.ValidatorAddr)]
				val.Minus = val.Minus.Add(delegation.Shares)
				currValidators[string(delegation.ValidatorAddr)] = val

				validatorPower := val.ValidatorInfo.EquivalentBondedShares(pool)
				delegatorShare := delegation.Shares.Quo(val.ValidatorInfo.DelegatorShares)
				votingPower := validatorPower.Mul(delegatorShare)
				results[VoteOptionToString(vote.Option)] = results[VoteOptionToString(vote.Option)].Add(votingPower)
				totalVotingPower = totalVotingPower.Add(votingPower)
			}
		}

		keeper.deleteVote(ctx, vote.ProposalID, vote.Voter)
	}
	votesIterator.Close()

	// Iterate over the validators again to tally their voting power and see who didn't vote
	nonVoting = []sdk.Address{}
	for _, val := range currValidators {
		if val.Vote == OptionEmpty {
			nonVoting = append(nonVoting, val.ValidatorInfo.Owner)
			continue
		}
		validatorPower := val.ValidatorInfo.EquivalentBondedShares(pool)
		sharesAfterMinus := val.ValidatorInfo.DelegatorShares.Sub(val.Minus)
		percentAfterMinus := sharesAfterMinus.Quo(val.ValidatorInfo.DelegatorShares)
		votingPower := validatorPower.Mul(percentAfterMinus)

		results[VoteOptionToString(val.Vote)] = results[VoteOptionToString(val.Vote)].Add(votingPower)
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	tallyingProcedure := keeper.GetTallyingProcedure(ctx)

	// If no one votes, proposal fails
	if totalVotingPower.Sub(results[VoteOptionToString(OptionAbstain)]).Equal(sdk.ZeroRat()) {
		return false, nonVoting
	}
	// If more than 1/3 of voters veto, proposal fails
	if results[VoteOptionToString(OptionNoWithVeto)].Quo(totalVotingPower).GT(tallyingProcedure.Veto) {
		return false, nonVoting
	}
	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[VoteOptionToString(OptionYes)].Quo(totalVotingPower.Sub(results[VoteOptionToString(OptionAbstain)])).GT(tallyingProcedure.Threshold) {
		return true, nonVoting
	}
	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, nonVoting
}
