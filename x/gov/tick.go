package gov

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) (tags sdk.Tags, nonVotingVals []sdk.Address) {

	tags = sdk.NewTags()

	// Delete proposals that haven't met minDeposit
	for shouldPopInactiveProposalQueue(ctx, keeper) {
		inactiveProposal := keeper.InactiveProposalQueuePop(ctx)
		if inactiveProposal.GetStatus() == StatusDepositPeriod {
			proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(inactiveProposal.GetProposalID())
			keeper.DeleteProposal(ctx, inactiveProposal)
			tags.AppendTag("action", []byte("proposalDropped"))
			tags.AppendTag("proposalId", proposalIDBytes)
		}
	}

	var passes bool

	// Check if earliest Active Proposal ended voting period yet
	for shouldPopActiveProposalQueue(ctx, keeper) {
		activeProposal := keeper.ActiveProposalQueuePop(ctx)

		if ctx.BlockHeight() >= activeProposal.GetVotingStartBlock()+keeper.GetVotingProcedure(ctx).VotingPeriod {
			passes, nonVotingVals = tally(ctx, keeper, activeProposal)
			proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(activeProposal.GetProposalID())
			if passes {
				keeper.RefundDeposits(ctx, activeProposal.GetProposalID())
				activeProposal.SetStatus(StatusPassed)
				tags.AppendTag("action", []byte("proposalPassed"))
				tags.AppendTag("proposalId", proposalIDBytes)
			} else {
				keeper.DeleteDeposits(ctx, activeProposal.GetProposalID())
				activeProposal.SetStatus(StatusRejected)
				tags.AppendTag("action", []byte("proposalRejected"))
				tags.AppendTag("proposalId", proposalIDBytes)
			}

			keeper.SetProposal(ctx, activeProposal)
		}
	}

	return tags, nonVotingVals
}

func tally(ctx sdk.Context, keeper Keeper, proposal Proposal) (passes bool, nonVoting []sdk.Address) {

	results := make(map[string]sdk.Rat)
	results[OptionYes] = sdk.ZeroRat()
	results[OptionAbstain] = sdk.ZeroRat()
	results[OptionNo] = sdk.ZeroRat()
	results[OptionNoWithVeto] = sdk.ZeroRat()

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
				results[vote.Option] = results[vote.Option].Add(votingPower)
				totalVotingPower = totalVotingPower.Add(votingPower)
			}
		}

		keeper.deleteVote(ctx, vote.ProposalID, vote.Voter)
	}
	votesIterator.Close()

	// Iterate over the validators again to tally their voting power and see who didn't vote
	nonVoting = []sdk.Address{}
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			nonVoting = append(nonVoting, val.ValidatorInfo.Owner)
			continue
		}
		validatorPower := val.ValidatorInfo.EquivalentBondedShares(pool)
		sharesAfterMinus := val.ValidatorInfo.DelegatorShares.Sub(val.Minus)
		percentAfterMinus := sharesAfterMinus.Quo(val.ValidatorInfo.DelegatorShares)
		votingPower := validatorPower.Mul(percentAfterMinus)

		results[val.Vote] = results[val.Vote].Add(votingPower)
		totalVotingPower = totalVotingPower.Add(votingPower)
	}

	tallyingProcedure := keeper.GetTallyingProcedure(ctx)

	// If no one votes, proposal fails
	if totalVotingPower.Sub(results[OptionAbstain]).Equal(sdk.ZeroRat()) {
		return false, nonVoting
	}
	// If more than 1/3 of voters veto, proposal fails
	if results[OptionNoWithVeto].Quo(totalVotingPower).GT(tallyingProcedure.Veto) {
		return false, nonVoting
	}
	// If more than 1/2 of non-abstaining voters vote Yes, proposal passes
	if results[OptionYes].Quo(totalVotingPower.Sub(results[OptionAbstain])).GT(tallyingProcedure.Threshold) {
		return true, nonVoting
	}
	// If more than 1/2 of non-abstaining voters vote No, proposal fails
	return false, nonVoting
}

func shouldPopInactiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	depositProcedure := keeper.GetDepositProcedure(ctx)
	peekProposal := keeper.InactiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if peekProposal.GetStatus() != StatusDepositPeriod {
		return true
	} else if ctx.BlockHeight() >= peekProposal.GetSubmitBlock()+depositProcedure.MaxDepositPeriod {
		return true
	}
	return false
}

func shouldPopActiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	votingProcedure := keeper.GetVotingProcedure(ctx)
	peekProposal := keeper.ActiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if ctx.BlockHeight() >= peekProposal.GetVotingStartBlock()+votingProcedure.VotingPeriod {
		return true
	}
	return false
}
