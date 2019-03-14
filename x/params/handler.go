package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

func NewHandler(k ProposalKeeper) sdk.Handler {
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

func handleMsgSubmitParameterChangeProposal(ctx sdk.Context, k ProposalKeeper, msg MsgSubmitParameterChangeProposal) sdk.Result {
	return proposal.HandleSubmit(ctx, k.cdc, k.proposal, ProposalChangeProto(msg.Changes), msg.SubmitForm)
}

func NewProposalHandler(k ProposalKeeper) proposal.Handler {
	return func(ctx sdk.Context, p proposal.Content) sdk.Error {
		switch p := p.(type) {
		case ProposalChange:
			return handleProposalChange(ctx, k, p)
		default:
			// XXX
			return nil
		}
	}
}

func handleProposalChange(ctx sdk.Context, k ProposalKeeper, p ProposalChange) (err sdk.Error) {
	for _, c := range p.Changes {
		s, ok := k.GetSubspace(c.Space)
		if !ok {
			return ErrUnknownSubspace(k.codespace, c.Space)
		}
		var rawerr error
		if len(c.Subkey) == 0 {
			rawerr = s.SetRaw(ctx, c.Key, c.Value)
		} else {
			rawerr = s.SetRawWithSubkey(ctx, c.Key, c.Subkey, c.Value)
		}

		if rawerr != nil {
			return ErrSettingParameter(k.codespace, c.Key, c.Subkey, c.Value, rawerr.Error())
		}
	}

	return
}
