package params

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// NewParamChangeProposalHandler creates a new governance Handler for a ParamChangeProposal
func NewParamChangeProposalHandler(k Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case ParameterChangeProposal:
			return handleParameterChangeProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized param proposal content type: %T", c)
		}
	}
}

func handleParameterChangeProposal(ctx sdk.Context, k Keeper, p ParameterChangeProposal) error {
	for _, c := range p.Changes {
		ss, ok := k.GetSubspace(c.Subspace)
		if !ok {
			return sdkerrors.Wrap(ErrUnknownSubspace, c.Subspace)
		}

		k.Logger(ctx).Info(
			fmt.Sprintf("attempt to set new parameter value; key: %s, value: %s", c.Key, c.Value),
		)

		if err := ss.Update(ctx, []byte(c.Key), []byte(c.Value)); err != nil {
			return sdkerrors.Wrapf(ErrSettingParameter, "key: %s, value: %s, err: %s", c.Key, c.Value, err.Error())
		}
	}

	return nil
}
