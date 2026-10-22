package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// ValidateInitialDeposit is a helper function used only in deposit tests which returns the same
// functionality of validateInitialDeposit private function.
func (k Keeper) ValidateInitialDeposit(ctx sdk.Context, initialDeposit sdk.Coins, expedited bool) error {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return err
	}

	return k.validateInitialDeposit(ctx, params, initialDeposit, expedited)
}

// NewMsgServerImpl and NewLegacyMsgServerImpl expose the unexported msg
// servers to the keeper_test package.
var (
	NewMsgServerImpl       = newMsgServerImpl
	NewLegacyMsgServerImpl = newLegacyMsgServerImpl
)
