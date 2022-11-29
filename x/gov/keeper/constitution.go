package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (keeper Keeper) GetConstitution(ctx sdk.Context) (constitution string) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte("constitution"))

	return string(bz)
}
