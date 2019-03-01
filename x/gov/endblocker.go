package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// Called every block, process inflation, update validator set
func EndBlocker(ctx sdk.Context, keeper Keeper) sdk.Tags {
	logger := ctx.Logger().With("module", "x/gov")
	resTags := sdk.NewTags()

	inactiveIterator := keeper.InactiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer inactiveIterator.Close()
	for ; inactiveIterator.Valid(); inactiveIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(inactiveIterator.Value(), &proposalID)
		inactiveProposal, ok := keeper.GetProposalProcess(ctx, proposalID)
		if !ok {
			panic("Endblocker cannot fail to GetProposalProcess")
		}

		keeper.DeleteProposalProcess(ctx, proposalID)
		keeper.DeleteDeposits(ctx, proposalID) // delete any associated deposits (burned)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tags.ActionProposalDropped)

		logger.Info(
			fmt.Sprintf("proposal %d (%s) didn't meet minimum deposit of %s (had only %s); deleted",
				inactiveProposal.ProposalID,
				inactiveProposal.GetTitle(),
				keeper.GetDepositParams(ctx).MinDeposit,
				inactiveProposal.TotalDeposit,
			),
		)
	}

	// fetch active proposals whose voting periods have ended (are passed the block time)
	activeIterator := keeper.ActiveProposalQueueIterator(ctx, ctx.BlockHeader().Time)
	defer activeIterator.Close()
	for ; activeIterator.Valid(); activeIterator.Next() {
		var proposalID uint64

		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(activeIterator.Value(), &proposalID)
		activeProposal, ok := keeper.GetProposalProcess(ctx, proposalID)
		if !ok {
			panic("EndBlocker cannot fail to GetProposalProcess")
		}
		passes, tallyResults := tally(ctx, keeper, activeProposal)

		var tagValue string
		if passes {
			keeper.RefundDeposits(ctx, activeProposal.ProposalID)
			activeProposal.Status = StatusPassed
			tagValue = tags.ActionProposalPassed
		} else {
			keeper.DeleteDeposits(ctx, activeProposal.ProposalID)
			activeProposal.Status = StatusRejected
			tagValue = tags.ActionProposalRejected
		}

		activeProposal.FinalTallyResult = tallyResults
		keeper.SetProposalProcess(ctx, activeProposal)
		keeper.RemoveFromActiveProposalQueue(ctx, activeProposal.VotingEndTime, activeProposal.ProposalID)

		logger.Info(
			fmt.Sprintf(
				"proposal %d (%s) tallied; passed: %v",
				activeProposal.ProposalID, activeProposal.GetTitle(), passes,
			),
		)

		resTags = resTags.AppendTag(tags.ProposalID, fmt.Sprintf("%d", proposalID))
		resTags = resTags.AppendTag(tags.ProposalResult, tagValue)
	}

	return resTags
}
