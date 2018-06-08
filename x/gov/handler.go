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

	_, _, err := keeper.ck.SubtractCoins(ctx, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{}
	}

	proposal := keeper.NewProposal(ctx, msg.Title, msg.Description, msg.ProposalType)

	err = keeper.AddDeposit(ctx, proposal.ProposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}

	tags := sdk.NewTags("action", []byte("submitProposal"), "proposer", msg.Proposer.Bytes(), "proposalId", []byte{byte(proposal.ProposalID)})
	return sdk.Result{
		Data: []byte{byte(proposal.ProposalID)},
		Tags: tags,
	}
}

// Handle MsgDeposit.
func handleMsgDeposit(ctx sdk.Context, keeper Keeper, msg MsgDeposit) sdk.Result {

	_, _, err := keeper.ck.SubtractCoins(ctx, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	if ctx.IsCheckTx() {
		return sdk.Result{} // TODO
	}

	err = keeper.AddDeposit(ctx, msg.ProposalID, msg.Depositer, msg.Amount)
	if err != nil {
		return err.Result()
	}

	// TODO: Add tag for if voting period started
	tags := sdk.NewTags("action", []byte("deposit"), "depositer", msg.Depositer.Bytes(), "proposalId", []byte{byte(msg.ProposalID)})
	return sdk.Result{
		Tags: tags,
	}
}

// Handle SendMsg.
func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {

	if ctx.IsCheckTx() {
		proposal := keeper.GetProposal(ctx, msg.ProposalID)

		if proposal == nil {
			return ErrUnknownProposal(msg.ProposalID).Result()
		}
		if (proposal.Status != "Pending") && (proposal.Status != "Active") {
			return ErrAlreadyFinishedProposal(msg.ProposalID).Result()
		}

		return sdk.Result{} // TODO
	}

	keeper.AddVote(ctx, msg.ProposalID, msg.Voter, msg.Option)

	tags := sdk.NewTags("action", []byte("vote"), "voter", msg.Voter.Bytes(), "proposalId", []byte{byte(msg.ProposalID)})
	return sdk.Result{
		Tags: tags,
	}
}
