package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// SetAppVersion sets the app version to version. Used for testing the upgrade keeper.
func (k *Keeper) SetAppVersion(ctx sdk.Context, version uint64) error {
	return k.versionManager.SetAppVersion(ctx, version)
}
