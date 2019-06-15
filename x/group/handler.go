package group

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "data" type messages.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgCreateGroup:
			return handleMsgCreateGroup(ctx, keeper, msg)
		case MsgCreateProposal:
			return handleMsgCreateProposal(ctx, keeper, msg)
		case MsgVote:
			return handleMsgVote(ctx, keeper, msg)
		case MsgTryExecuteProposal:
			return handleMsgTryExecuteProposal(ctx, keeper, msg)
		case MsgWithdrawProposal:
			return handleMsgWithdrawProposal(ctx, keeper, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized data Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgCreateGroup(ctx sdk.Context, keeper Keeper, msg MsgCreateGroup) sdk.Result {
	id, err := keeper.CreateGroup(ctx, msg.Data)
	if err != nil {
		return sdk.Result{
			Code: sdk.CodeUnknownAddress,
			Log:  err.Error(),
		}
	}
	return sdk.Result{
		Tags: sdk.NewTags("group.id", []byte(id.String())),
	}
}


func handleMsgCreateProposal(ctx sdk.Context, keeper Keeper, msg MsgCreateProposal) sdk.Result {
	id, res := keeper.Propose(ctx, msg.Proposer, msg.Action)
	if res.Code != sdk.CodeOK {
		return res
	}
	if msg.Exec {
		res := keeper.TryExecute(ctx, id)
		if res.Code == sdk.CodeOK {
			return res
		}
	}
	return res
}

func handleMsgVote(ctx sdk.Context, keeper Keeper, msg MsgVote) sdk.Result {
	return keeper.Vote(ctx, msg.ProposalId, msg.Voter, msg.Vote)
}

func handleMsgTryExecuteProposal(ctx sdk.Context, keeper Keeper, msg MsgTryExecuteProposal) sdk.Result {
	return keeper.TryExecute(ctx, msg.ProposalId)
}

func handleMsgWithdrawProposal(ctx sdk.Context, keeper Keeper, msg MsgWithdrawProposal) sdk.Result {
	return keeper.Withdraw(ctx, msg.ProposalId, msg.Proposer)
}
