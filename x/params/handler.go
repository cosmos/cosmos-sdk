package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
	"github.com/cosmos/cosmos-sdk/x/params/tags"
)

func NewHandler(k ProposalKeeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgSubmitProposal:
			return handleMsgSubmitProposal(ctx, k, msg)
		default:
			errMsg := fmt.Sprintf("Unrecognized gov msg type: %T", msg)
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

func handleMsgSubmitProposal(ctx sdk.Context, k ProposalKeeper, msg MsgSubmitProposal) sdk.Result {
	return proposal.HandleSubmit(ctx, k.proposal, msg.Content, msg.Proposer, msg.InitialDeposit, tags.TxCategory)
}

func NewProposalHandler(k ProposalKeeper) proposal.Handler {
	return func(ctx sdk.Context, p proposal.Content) sdk.Error {
		switch p := p.(type) {
		case ChangeProposal:
			return handleChangeProposal(ctx, k, p)
		default:
			errMsg := fmt.Sprintf("Unrecognized gov proposal type: %T", p)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleChangeProposal(ctx sdk.Context, k ProposalKeeper, p ChangeProposal) sdk.Error {
	for _, c := range p.Changes {
		s, ok := k.GetSubspace(c.Space)
		if !ok {
			return ErrUnknownSubspace(k.codespace, c.Space)
		}
		var err error
		if len(c.Subkey) == 0 {
			err = s.SetRaw(ctx, c.Key, c.Value)
		} else {
			err = s.SetRawWithSubkey(ctx, c.Key, c.Subkey, c.Value)
		}

		if err != nil {
			return ErrSettingParameter(k.codespace, c.Key, c.Subkey, c.Value, err.Error())
		}
	}

	return nil
}
