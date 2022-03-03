package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	updateTallyOfVPEndProposals(ctx, k)
}

func updateTallyOfVPEndProposals(ctx sdk.Context, k keeper.Keeper) error {
	k.IterateVPEndProposals(ctx, sdk.FormatTimeBytes(ctx.BlockTime()), func(proposal group.Proposal) (stop bool) {

		policyInfo, err := k.GetGroupPolicyInfo(ctx, proposal.Address)
		if err != nil {
			return true
		}

		tallyRes, err := k.Tally(ctx, proposal, policyInfo.GroupId)
		if err != nil {
			return true
		}

		proposal.FinalTallyResult = tallyRes
		err = k.UpdateProposal(ctx, proposal)

		return err != nil
	})

	return nil
}
