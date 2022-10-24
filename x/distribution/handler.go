package distribution

import (
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	sdkerrors "github.com/pointnetwork/cosmos-point-sdk/types/errors"
	"github.com/pointnetwork/cosmos-point-sdk/x/distribution/keeper"
	"github.com/pointnetwork/cosmos-point-sdk/x/distribution/types"
	govtypes "github.com/pointnetwork/cosmos-point-sdk/x/gov/types/v1beta1"
)

func NewCommunityPoolSpendProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CommunityPoolSpendProposal:
			return keeper.HandleCommunityPoolSpendProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized distr proposal content type: %T", c)
		}
	}
}
