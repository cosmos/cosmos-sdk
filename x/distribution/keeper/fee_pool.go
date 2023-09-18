package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const poolModuleName = "protocol-pool"

// DistributeFromFeePool distributes funds from the pool module account to
// a receiver address
func (k Keeper) DistributeFromFeePool(ctx context.Context, amount sdk.Coins, receiveAddr sdk.AccAddress) error {
	err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, poolModuleName, receiveAddr, amount)
	if err != nil {
		return err
	}
	return nil
}
