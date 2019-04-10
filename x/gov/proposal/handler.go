package proposal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// HandleSubmit submits content and does initial deposit with provided SubmitForm
func HandleSubmit(ctx sdk.Context, k Keeper,
	content Content, proposer sdk.AccAddress, initialDeposit sdk.Coins,
	tagTxCategory string) sdk.Result {

	proposalID, err := k.SubmitProposal(ctx, content)
	if err != nil {
		return err.Result()
	}
	proposalIDStr := fmt.Sprintf("%d", proposalID)

	err, votingStarted := k.AddDeposit(ctx, proposalID, proposer, initialDeposit)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Proposer, proposer.String(),
		tags.ProposalID, proposalIDStr,
		tags.Category, tagTxCategory,
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Data: internalCdc.MustMarshalBinaryLengthPrefixed(proposalID),
		Tags: resTags,
	}
}
