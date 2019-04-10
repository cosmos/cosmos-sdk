package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// DistributeFeePool distributes funds from the the community pool to a receiver address
func (k Keeper) DistributeFeePool(ctx sdk.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) sdk.Error {
	feePool := k.GetFeePool(ctx)

	poolTruncated, _ := feePool.CommunityPool.TruncateDecimal()
	if !poolTruncated.IsAllGTE(amount) {
		return types.ErrBadDistribution(k.codespace)
	}

	feePool.CommunityPool.Sub(sdk.NewDecCoins(amount))
	_, err := k.bankKeeper.AddCoins(ctx, receiveAddr, amount)
	if err != nil {
		return err
	}

	k.SetFeePool(ctx, feePool)
	return nil
}
