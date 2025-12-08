package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// ValidateInitialDeposit is a helper function used only in deposit tests which returns the same
// functionality of validateInitialDeposit private function.
func (k Keeper) ValidateInitialDeposit(ctx sdk.Context, initialDeposit sdk.Coins, expedited bool) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	return k.validateInitialDeposit(ctx, params, initialDeposit, expedited)
}

// ValidateInitialDepositWithParams exposes validateInitialDeposit for testing.
func (k Keeper) ValidateInitialDepositWithParams(ctx sdk.Context, params v1.Params, initialDeposit sdk.Coins, expedited bool) error {
	return k.validateInitialDeposit(ctx, params, initialDeposit, expedited)
}
