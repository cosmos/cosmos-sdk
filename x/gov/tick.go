package gov

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) {

	peekProposal := keeper.ProposalQueuePeek(ctx)

	keeper.GetVotingProcedure()

	if ctx.BlockHeight() > peekProposal.VotingStartBlock+keeper.GetVotingProcedure().VotingPeriod {
		passes, _ := tally(ctx, keeper, peekProposal)
		if passes {
			keeper.RefundDeposits(ctx, peekProposal.ProposalID)
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

	totalVotingPower := sdk.ZeroRat()
	currValidators := make(map[string]validatorGovInfo)
	for _, val := range keeper.sk.GetValidatorsBonded(ctx) {
		currValidators[addressToString(val.Owner)] = validatorGovInfo{
			ValidatorInfo: val,
			Minus:         sdk.ZeroRat(),
		}
	}

	votesIterator := keeper.GetVotes(ctx, proposal.ProposalID)
	for ; votesIterator.Valid(); votesIterator.Next() {
		vote := &Vote{}
		keeper.cdc.MustUnmarshalBinary(votesIterator.Value(), vote)

		if val, ok := currValidators[addressToString(vote.Voter)]; ok {
			val.Vote = vote.Option
		} else {
			for _, delegation := range keeper.sk.GetDelegations(ctx, vote.Voter, math.MaxInt16) {
				val := currValidators[addressToString(delegation.ValidatorAddr)]
				val.Minus = val.Minus.Add(delegation.Shares)

				votingPower := val.ValidatorInfo.PoolShares.Amount.Mul(delegation.Shares)
				results[vote.Option] = results[vote.Option].Add(votingPower)
				totalVotingPower = totalVotingPower.Add(votingPower)
			}
		}
	}
	votesIterator.Close()

	nonVoting = []sdk.Address{}
	for _, val := range currValidators {
		if len(val.Vote) == 0 {
			nonVoting = append(nonVoting, val.ValidatorInfo.Owner)
		} else {
			sharesAfterMinus := val.ValidatorInfo.DelegatorShares.Sub(val.Minus)
			votingPower := sharesAfterMinus.Mul(val.ValidatorInfo.PoolShares.Amount)

			results[val.Vote] = results[val.Vote].Add(votingPower)
			totalVotingPower = totalVotingPower.Add(votingPower)
		}
	}

	tallyingProcedure := keeper.GetTallyingProcedure()

	if results["NoWithVeto"].Quo(totalVotingPower).GT(tallyingProcedure.Veto) {
		return false, nonVoting
	} else if results["Yes"].Quo(totalVotingPower.Sub(results["Abstain"])).GT(tallyingProcedure.Threshold) {
		return true, nonVoting
	} else {
		return false, nonVoting
	}
}

func addressToString(addr sdk.Address) string {
	return fmt.Sprintf("%s", addr.Bytes())
}

// type Delegation struct {
// 	DelegatorAddr sdk.Address `json:"delegator_addr"`
// 	ValidatorAddr sdk.Address `json:"validator_addr"`
// 	Shares        sdk.Rat     `json:"shares"`
// 	Height        int64       `json:"height"` // Last height bond updated
// }

/*
// Procedure around Tallying votes in governance
type TallyingProcedure struct {
	Threshold         sdk.Rat `json:"threshold"`          //  Minimum propotion of Yes votes for proposal to pass. Initial value: 0.5
	Veto              sdk.Rat `json:"veto"`               //  Minimum value of Veto votes to Total votes ratio for proposal to be vetoed. Initial value: 1/3
	GovernancePenalty sdk.Rat `json:"governance_penalty"` //  Penalty if validator does not vote
}



func NewBeginBlocker(gm governanceMapper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		proposal := gm.ProposalQueuePeek(ctx)
		if proposal == nil {
			return abci.ResponseBeginBlock{} // TODO
		}

		// Don't want to do urgent for now

		// // Urgent proposal accepted
		// if proposal.Votes.YesVotes/proposal.InitTotalVotingPower >= 2/3 {
		// 	gm.PopProposalQueue(ctx)
		// 	refund(ctx, gm, proposalID, proposal)
		// 	return checkProposal()
		// }

		// Proposal reached the end of the voting period
		if ctx.BlockHeight() == proposal.VotingStartBlock+proposal.Procedure.VotingPeriod {
			gm.ProposalQueuePop(ctx)

			// Slash validators if not voted
			for _, validatorGovInfo := range proposal.ValidatorGovInfos {
				if validatorOption.LastVoteWeight < 0 {
					// TODO: SLASH MWAHAHAHAHAHA
				}
			}

			//Proposal was accepted
			nonAbstainTotal := proposal.Votes.YesVotes + proposal.Votes.NoVotes + proposal.Votes.NoWithVetoVotes
			if proposal.YesVotes/nonAbstainTotal > 0.5 && proposal.NoWithVetoVotes/nonAbstainTotal < 1/3 { // TODO: Deal with decimals

				//  TODO:  Act upon accepting of proposal

				// Refund deposits
				for _, deposit := range proposal.Deposits {
					gm.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
					if err != nil {
						panic("should not happen")
					}
				}

				// check next proposal recursively
				checkProposal()
			}

			//  TODO: Prune proposal
		}
		return abci.ResponseBeginBlock{}
	}
}

*/
