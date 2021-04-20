package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// HandleCommunityPoolSpendProposal is a handler for executing a passed community spend proposal
func HandleCommunityPoolSpendProposal(ctx sdk.Context, k Keeper, p *types.CommunityPoolSpendProposal) error {
	if k.blockedAddrs[p.Recipient] {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive external funds", p.Recipient)
	}

	recipient, addrErr := sdk.AccAddressFromBech32(p.Recipient)
	if addrErr != nil {
		return addrErr
	}

	err := k.DistributeFromFeePool(ctx, p.Amount, recipient)
	if err != nil {
		return err
	}

	logger := k.Logger(ctx)
	logger.Info("transferred from the community pool to recipient", "amount", p.Amount.String(), "recipient", p.Recipient)

	return nil
}
