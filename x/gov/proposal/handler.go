package proposal

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// HandleSubmit submits content and does initial deposit with provided SubmitForm
func HandleSubmit(ctx sdk.Context, cdc *codec.Codec, k Keeper, proto Proto, form SubmitForm) sdk.Result {
	content := proto(form.Title, form.Description)
	proposalID, err := k.SubmitProposal(ctx, content)
	if err != nil {
		return err.Result()
	}
	proposalIDStr := fmt.Sprintf("%d", proposalID)

	err, votingStarted := k.AddDeposit(ctx, proposalID, form.Proposer, form.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	resTags := sdk.NewTags(
		tags.Proposer, []byte(form.Proposer.String()),
		tags.ProposalID, proposalIDStr,
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Data: cdc.MustMarshalBinaryLengthPrefixed(proposalID),
		Tags: resTags,
	}
}
