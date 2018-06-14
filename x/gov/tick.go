package gov

import (
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) {

	// Delete proposals that haven't met minDeposit

	for shouldPopInactiveProposalQueue(ctx, keeper) {
		inactiveProposal := keeper.InactiveProposalQueuePop(ctx)
		if inactiveProposal.Status == StatusDepositPeriod {
			keeper.DeleteProposal(ctx, inactiveProposal)
		}
	}

	// Check if earliest Active Proposal ended voting period yet

	for shouldPopActiveProposalQueue(ctx, keeper) {
		activeProposal := keeper.ActiveProposalQueuePop(ctx)

		if ctx.BlockHeight() >= activeProposal.VotingStartBlock+keeper.GetVotingProcedure(ctx).VotingPeriod {
			passes, _ := tally(ctx, keeper, activeProposal)
			if passes {
				keeper.RefundDeposits(ctx, activeProposal.ProposalID)
				activeProposal.Status = StatusPassed
			} else {
				keeper.DeleteDeposits(ctx, activeProposal.ProposalID)
				activeProposal.Status = StatusRejected
			}

			keeper.SetProposal(ctx, activeProposal)
		}
	}

	return
}

func tally(ctx sdk.Context, keeper Keeper, proposal *Proposal) (passes bool, nonVoting []sdk.Address) {

	results := make(map[string]sdk.Rat)
	results["Yes"] = sdk.ZeroRat()
	results["Abstain"] = sdk.ZeroRat()
	results["No"] = sdk.ZeroRat()
	results["NoWithVeto"] = sdk.ZeroRat()

	pool := keeper.sk.GetPool(ctx)

	totalVotingPower := sdk.ZeroRat()
	currValidators := make(map[string]validatorGovInfo)
	for _, val := range keeper.sk.GetValidatorsBonded(ctx) {
		currValidators[string(val.Owner)] = validatorGovInfo{
			ValidatorInfo: val,
			Minus:         sdk.ZeroRat(),
		}
	}

	votesIterator := keeper.GetVotes(ctx, proposal.ProposalID)
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)

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

	nonVoting = []sdk.Address{}
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			nonVoting = append(nonVoting, val.ValidatorInfo.Owner)
		} else {
			validatorPower := val.ValidatorInfo.EquivalentBondedShares(pool)
			sharesAfterMinus := val.ValidatorInfo.DelegatorShares.Sub(val.Minus)
			percentAfterMinus := sharesAfterMinus.Quo(val.ValidatorInfo.DelegatorShares)
			votingPower := validatorPower.Mul(percentAfterMinus)

			results[val.Vote] = results[val.Vote].Add(votingPower)
			totalVotingPower = totalVotingPower.Add(votingPower)
		}
	}

	tallyingProcedure := keeper.GetTallyingProcedure(ctx)

	if totalVotingPower.Sub(results["Abstain"]).Equal(sdk.ZeroRat()) {
		return false, nonVoting
	} else if results["NoWithVeto"].Quo(totalVotingPower).GT(tallyingProcedure.Veto) {
		return false, nonVoting
	} else if results["Yes"].Quo(totalVotingPower.Sub(results["Abstain"])).GT(tallyingProcedure.Threshold) {
		return true, nonVoting
	}
	return false, nonVoting
}

func shouldPopInactiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	depositProcedure := keeper.GetDepositProcedure(ctx)
	peekProposal := keeper.InactiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if peekProposal.Status != StatusDepositPeriod {
		return true
	} else if ctx.BlockHeight() >= peekProposal.SubmitBlock+depositProcedure.MaxDepositPeriod {
		return true
	}
	return false
}

func shouldPopActiveProposalQueue(ctx sdk.Context, keeper Keeper) bool {
	votingProcedure := keeper.GetVotingProcedure(ctx)
	peekProposal := keeper.ActiveProposalQueuePeek(ctx)

	if peekProposal == nil {
		return false
	} else if ctx.BlockHeight() >= peekProposal.VotingStartBlock+votingProcedure.VotingPeriod {
		return true
	}
	return false
}
