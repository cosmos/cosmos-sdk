package keeper

import "context"

// EndBlock returns an endblocker for the x/feemarket module. The endblocker
// is responsible for updating the state of the fee market based on the
// AIMD learning rate adjustment algorithm.
func (k *Keeper) EndBlock(ctx context.Context) error {
	return k.UpdateFeeMarket(ctx)
}
