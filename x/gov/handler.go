package gov

import (
	"reflect"

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
			errMsg := "Unrecognized gov Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgSubmitProposal.
func handleMsgSubmitProposal(ctx sdk.Context, keeper Keeper, msg MsgSubmitProposal) sdk.Result {

	proposal := keeper.NewProposal(ctx, msg.Title, msg.Description, msg.ProposalType)

	err := keeper.AddDeposit(ctx, proposal.ProposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes, _ := keeper.cdc.MarshalBinaryBare(proposal.ProposalID)

	tags := sdk.NewTags("action", []byte("submitProposal"), "proposer", msg.Proposer.Bytes(), "proposalId", proposalIDBytes)
	return sdk.Result{
		Data: proposalIDBytes,
		Tags: tags,
	}
}

// Handle MsgDeposit.
func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	err := keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes, _ := keeper.cdc.MarshalBinaryBare(msg.ProposalID)

	// TODO: Add tag for if voting period started
	tags := sdk.NewTags("action", []byte("deposit"), "depositer", msg.Depositer.Bytes(), "proposalId", proposalIDBytes)
	return sdk.Result{
		Tags: tags,
	}
}

// Handle SendMsg.
func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {

	err := keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)
	if err != nil {
		return err.Result()
	}

	proposalIDBytes, _ := keeper.cdc.MarshalBinaryBare(msg.ProposalID)

	tags := sdk.NewTags("action", []byte("vote"), "voter", msg.Voter.Bytes(), "proposalId", proposalIDBytes)
	return sdk.Result{
		Tags: tags,
	}
}
