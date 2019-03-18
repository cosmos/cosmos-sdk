package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// initialize starting info for a new delegation
func (k Keeper) DistributeFeePool(ctx sdk.Context, amount sdk.Coin, receiveAddr sdk.AccAddress) error {
	feePool := k.GetFeePool(ctx)

	existingCoin := feePool.CommunityPool.AmountOf(amount.Denom)
	if existingCoin.LT(sdk.NewDecFromInt(amount.Amount)) {
		return errors.New("community pool does not have sufficient coins to distribute")
	}

	feePool.CommunityPool.Sub(sdk.NewDecCoins(sdk.NewCoins(amount)))
	k.bankKeeper.AddCoins(ctx, receiveAddr, sdk.Coins{amount})

	k.SetFeePool(ctx, feePool)
	return nil
}
