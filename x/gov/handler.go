package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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
			errMsg := "Unrecognized gov msg type"
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {

	proposal := keeper.NewTextProposal(ctx, msg.Title, msg.Description, msg.ProposalType)

	err, votingStarted := keeper.AddDeposit(ctx, proposal.GetProposalID(), msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(proposal.GetProposalID())

	tags := sdk.NewTags(
		"action", []byte("submitProposal"),
		"proposer", []byte(msg.Proposer.String()),
		"proposalId", proposalIDBytes,
	)

	if votingStarted {
		tags.AppendTag("votingPeriodStart", proposalIDBytes)
	}

	return sdk.Result{
		Data: proposalIDBytes,
		Tags: tags,
	}
}

func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	err, votingStarted := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	// TODO: Add tag for if voting period started
	tags := sdk.NewTags(
		"action", []byte("deposit"),
		"depositer", []byte(msg.Depositer.String()),
		"proposalId", proposalIDBytes,
	)

	if votingStarted {
		tags.AppendTag("votingPeriodStart", proposalIDBytes)
	}

	return sdk.Result{
		Tags: tags,
	}
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {

	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes := keeper.cdc.MustMarshalBinaryBare(msg.ProposalID)

	tags := sdk.NewTags(
		"action", []byte("vote"),
		"voter", []byte(msg.Voter.String()),
		"proposalId", proposalIDBytes,
	)
	return sdk.Result{
		Tags: tags,
	}
}
