package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// ValidateInitialDeposit is a helper function used only in deposit tests which returns the same
// functionality of validateInitialDeposit private function.
func (keeper Keeper) ValidateInitialDeposit(ctx sdk.Context, initialDeposit sdk.Coins) error {
	params, err := keeper.Params.Get(ctx)
	if err != nil {
		return err
	}

	return keeper.validateInitialDeposit(ctx, params, initialDeposit)
}
