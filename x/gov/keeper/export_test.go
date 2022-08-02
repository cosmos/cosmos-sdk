package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) ValidateInitialDeposit(ctx sdk.Context, initialDeposit sdk.Coins) error {
	return k.validateInitialDeposit(ctx, initialDeposit)
}
