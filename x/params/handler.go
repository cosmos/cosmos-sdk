package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/proposal"
)

func NewProposalHandler(k Keeper) proposal.Handler {
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

func handleChangeProposal(ctx sdk.Context, k Keeper, p ChangeProposal) sdk.Error {
	for _, c := range p.Changes {
		s, ok := k.GetSubspace(c.Space)
		if !ok {
			return ErrUnknownSubspace(k.codespace, c.Space)
		}
		var err error
		if len(c.Subkey) == 0 {
			err = s.SetRaw(ctx, []byte(c.Key), c.Value)
		} else {
			err = s.SetRawWithSubkey(ctx, []byte(c.Key), c.Subkey, c.Value)
		}

		if err != nil {
			return ErrSettingParameter(k.codespace, []byte(c.Key), c.Subkey, c.Value, err.Error())
		}
	}

	return nil
}
