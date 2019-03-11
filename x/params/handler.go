package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSubmitParameterChangeProposal:
			return handleMsgSubmitParameterChangeProposal(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized gov msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitParameterChangeProposal(ctx sdk.Context, k Keeper, msg MsgSubmitParameterChangeProposal) sdk.Result {
	// XXX: generalize
	// we can make gov import params.subspace, gov can export helper functions
	// which params can import and use without import cycle

	content := NewProposalChange(msg.Title, msg.Description, msg.Space, msg.Changes)
	proposalID, err := k.pk.SubmitProposal(ctx, content)
	if err != nil {
		return err.Result()
	}
	err, _ = k.pk.AddDeposit(ctx, proposalID, msg.Proposer, msg.InitialDeposit)
	if err != nil {
		return err.Result()
	}
	/*
		resTags := sdk.NewTags(

		)
	*/
	return sdk.Result{}
}

func NewProposalHandler(k Keeper) sdk.ProposalHandler {
	return func(ctx sdk.Context, p sdk.ProposalContent) sdk.Error {
		switch p := p.(type) {
		case ProposalChange:
			return handleProposalChange(ctx, k, p)
		default:
			// XXX
			return nil
		}
	}
}

func handleProposalChange(ctx sdk.Context, k Keeper, p ProposalChange) sdk.Error {
	s, ok := k.GetSubspace(p.Space)
	if !ok {
		// XXX
		return nil
	}

	for _, c := range p.Changes {
		if len(c.Subkey) == 0 {
			s.Set(ctx, c.Key, c.Value)
		} else {
			s.SetWithSubkey(ctx, c.Key, c.Subkey, c.Value)
		}
	}

	return nil
}
