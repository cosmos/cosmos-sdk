package gov

import (
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

func NewBeginBlocker(gm Keeper) sdk.BeginBlocker {
	return func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		proposal := gm.ProposalQueuePeek(ctx)
		if proposal == nil {
			return abci.ResponseBeginBlock{}
		}

		// Don't want to do urgent for now
		passV := types.NewRat(proposal.YesVotes, proposal.TotalVotingPower)
		// // Urgent proposal accepted
		if passV.GT(proposal.Procedure.FastPass) || passV.Equal(proposal.Procedure.FastPass) {

			ctx.Logger().Info("execute Proposal", "Proposal", proposal.ProposalID)

			gm.ProposalQueuePop(ctx)

			refund(ctx, proposal, gm)

			//TODO proposal.execute

			return abci.ResponseBeginBlock{}
		}

		// Proposal reached the end of the voting period
		if ctx.BlockHeight() >= proposal.VotingStartBlock+proposal.Procedure.VotingPeriod {
			gm.ProposalQueuePop(ctx)

			// Refund deposits
			refund(ctx, proposal, gm)

			//Slash validators if not voted
			slash(ctx, proposal.ValidatorGovInfos)

			//Proposal was accepted
			nonAbstainTotal := proposal.YesVotes + proposal.NoVotes + proposal.NoWithVetoVotes
			if nonAbstainTotal <= 0 {
				return abci.ResponseBeginBlock{}
			}
			yRat := types.NewRat(proposal.YesVotes, nonAbstainTotal)
			vetoRat := types.NewRat(proposal.NoWithVetoVotes, nonAbstainTotal)
			if yRat.GT(proposal.Procedure.Threshold) && vetoRat.LT(proposal.Procedure.Veto) {
				ctx.Logger().Info("Execute proposal", "proposal", proposal)
				//	TODO proposal.execute
				// check next proposal recursively
				//checkProposal()
			}
		}
		return abci.ResponseBeginBlock{}
	}
}

func refund(ctx sdk.Context, proposal *Proposal, govKeeper Keeper) {
	for _, deposit := range proposal.Deposits {
		ctx.Logger().Info("Execute Refund", "Depositer", deposit.Depositer, "Amount", deposit.Amount)
		_, _, err := govKeeper.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}
	}
}

func slash(ctx sdk.Context, validators []ValidatorGovInfo) {
	ctx.Logger().Info("Begin to Execute Slash")
	// Slash validators if not voted
	for _, validatorGovInfo := range validators {
		if validatorGovInfo.LastVoteWeight < 0 {
			// TODO: SLASH MWAHAHAHAHAHA
		}
	}
}
