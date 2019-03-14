package proposal

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// HandleSubmit submits content and does initial deposit with provided SubmitForm
func HandleSubmit(ctx sdk.Context, k Keeper,
	content Content, proposer sdk.AccAddress, initialDeposit sdk.Coins) sdk.Result {
	// XXX: ValidateSubmitProposalBasic is supposed to be called in each msg types
	// It it inefficient to do it here again?
	// I think we still should do, proposer can still skip CheckTx()
	err := ValidateMsgBasic(content.GetTitle(), content.GetDescription(), proposer, initialDeposit)
	if err != nil {
		return err.Result()
	}

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
		tags.Proposer, []byte(proposer.String()),
		tags.ProposalID, proposalIDStr,
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Data: internalCdc.MustMarshalBinaryLengthPrefixed(proposalID),
		Tags: resTags,
	}
}
