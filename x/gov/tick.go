package gov



import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/types"
)

func NewBeginBlocker(gm Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		proposal := gm.ProposalQueuePeek(ctx)
		if proposal == nil {
			return abci.ResponseBeginBlock{} // TODO
		}

		ctx.Logger().Info("gov","Proposal",proposal)

		// Don't want to do urgent for now
		passV := types.NewRat(proposal.YesVotes,proposal.TotalVotingPower)
		r := types.NewRat(2,3)

		// // Urgent proposal accepted
		if passV.GT(r) || passV.Equal(r){

			ctx.Logger().Info("execute Proposal","Proposal",proposal.ProposalID)

			gm.ProposalQueuePop(ctx)

			for _, deposit := range proposal.Deposits {
				ctx.Logger().Info("refund coins","Depositer",deposit.Depositer,"Amount",deposit.Amount)
				gm.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
				//if err != nil {
				//	panic("should not happen")
				//}
			}

			//refund(ctx, gm, proposalID, proposal)
			//return checkProposal()
		}

		// Proposal reached the end of the voting period
		if ctx.BlockHeight() == proposal.VotingStartBlock+proposal.Procedure.VotingPeriod {
			gm.ProposalQueuePop(ctx)

			// Slash validators if not voted
			for _, validatorGovInfo := range proposal.ValidatorGovInfos {
				if validatorGovInfo.LastVoteWeight < 0 {
					// TODO: SLASH MWAHAHAHAHAHA
				}
			}

			//Proposal was accepted
			nonAbstainTotal := proposal.YesVotes + proposal.NoVotes + proposal.NoWithVetoVotes
			yRat := types.NewRat(proposal.YesVotes,nonAbstainTotal)
			vetoRat := types.NewRat(proposal.NoWithVetoVotes,nonAbstainTotal)
			if yRat.GT(proposal.Procedure.Threshold) && vetoRat.LT(proposal.Procedure.Veto) { // TODO: Deal with decimals

				//  TODO:  Act upon accepting of proposal

				// Refund deposits
				for _, deposit := range proposal.Deposits {
					gm.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
					//if err != nil {
					//	panic("should not happen")
					//}
				}

				// check next proposal recursively
				//checkProposal()
			}

			//  TODO: Prune proposal
		}
		return abci.ResponseBeginBlock{}
	}
}
