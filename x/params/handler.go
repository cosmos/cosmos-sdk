package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func NewProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) sdk.Error {
		switch c := content.(type) {
		case ParameterChangeProposal:
			return handleParameterChangeProposal(ctx, k, c)

		default:
			errMsg := fmt.Sprintf("unrecognized param proposal content type: %T", c)
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

func handleParameterChangeProposal(ctx sdk.Context, k Keeper, p ParameterChangeProposal) sdk.Error {
	for _, c := range p.Changes {
		ss, ok := k.GetSubspace(c.Subspace)
		if !ok {
			return ErrUnknownSubspace(k.codespace, c.Subspace)
		}
		var err error
		if len(c.Subkey) == 0 {
			err = ss.SetRaw(ctx, []byte(c.Key), c.Value)
		} else {
			err = ss.SetRawWithSubkey(ctx, []byte(c.Key), c.Subkey, c.Value)
		}

		if err != nil {
			return ErrSettingParameter(k.codespace, []byte(c.Key), c.Subkey, c.Value, err.Error())
		}
	}

	return nil
}
