package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/tags"
)

// Handle all "gov" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		ctx = ctx.WithEventManager(sdk.NewEventManager())

		switch msg := msg.(type) {
		case MsgDeposit:
			return handleMsgDeposit(ctx, keeper, msg)

		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, keeper, msg)

		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("unrecognized gov message type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {
	proposal, err := keeper.SubmitProposal(ctx, msg.Content)
	if err != nil {
		return err.Result()
	}

	err, votingStarted := keeper.AddDeposit(ctx, proposal.ProposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			tags.SubmitProposal,
			sdk.NewAttribute(sdk.AttributeKeySender, msg.Proposer.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, tags.TxCategory),
		),
	})

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				tags.SubmitProposal,
				sdk.NewAttribute(tags.VotingPeriodStart, fmt.Sprintf("%d", proposal.ProposalID)),
			),
		)
	}

	return sdk.Result{
		Data:   keeper.cdc.MustMarshalBinaryLengthPrefixed(proposal.ProposalID),
		Events: ctx.EventManager().Events(),
	}
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {
	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositor, msg.Amount)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, tags.TxCategory),
		),
	)

	if votingStarted {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				tags.ProposalDeposit,
				sdk.NewAttribute(tags.VotingPeriodStart, fmt.Sprintf("%d", msg.ProposalID)),
			),
		)
	}

	return sdk.Result{Events: ctx.EventManager().Events()}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, tags.TxCategory),
		),
	)

	return sdk.Result{Events: ctx.EventManager().Events()}

}
