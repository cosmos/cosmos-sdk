package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DistributeFromFeePool distributes funds from the pool module account to
// a receiver address
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, "cosmos-pool", receiveAddr, amount)
	if err != nil {
		return err
	}
	return nil
}
