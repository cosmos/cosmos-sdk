package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// EndBlock returns an endblocker for the x/feemarket module. The endblocker
// is responsible for updating the state of the fee market based on the
// AIMD learning rate adjustment algorithm.
func (k *Keeper) EndBlock(ctx sdk.Context) error {
	return k.UpdateFeeMarket(ctx)
}
