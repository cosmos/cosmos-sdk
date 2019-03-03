package params

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
