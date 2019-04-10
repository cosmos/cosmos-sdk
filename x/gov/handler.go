package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized gov msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {
	return proposal.HandleSubmit(ctx, keeper, msg.Content, msg.Proposer, msg.InitialDeposit, tags.TxCategory)
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {
	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositor, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDStr := fmt.Sprintf("%d", msg.ProposalID)

	resTags := sdk.NewTags(
		tags.ProposalID, proposalIDStr,
		tags.Category, tags.TxCategory,
		tags.Depositor, msg.Depositor.String(),
	)

	if votingStarted {
		resTags = resTags.AppendTag(tags.VotingPeriodStart, proposalIDStr)
	}

	return sdk.Result{
		Tags: resTags,
	}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	proposalIDStr := fmt.Sprintf("%d", msg.ProposalID)

	return sdk.Result{
		Tags: sdk.NewTags(
			tags.ProposalID, proposalIDStr,
			tags.Category, tags.TxCategory,
			tags.Voter, msg.Voter.String(),
		),
	}
}

// ProposalHandler is a proposal.Handler that processes proposal.Content.
// It does nothing since both TextProposal and SoftUpgradeProposal
// does not directly effect the state.
func ProposalHandler(ctx sdk.Context, p proposal.Content) sdk.Error {
	switch p.(type) {
	case TextProposal, SoftwareUpgradeProposal:
		// Both proposal type does not effect on the state
		return nil
	default:
		errMsg := fmt.Sprintf("Unrecognized gov proposal type: %T", p)
		return sdk.ErrUnknownRequest(errMsg)
	}
}
