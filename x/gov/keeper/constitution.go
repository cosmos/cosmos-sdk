package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

// GetConstitution gets the chain's constitution.
func (keeper Keeper) GetConstitution(ctx context.Context) (string, error) {
	store := keeper.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.KeyConstitution)

	return string(bz), err
}

// GetConstitution sets the chain's constitution.
func (keeper Keeper) SetConstitution(ctx context.Context, constitution string) error {
	store := keeper.storeService.OpenKVStore(ctx)
	return store.Set(types.KeyConstitution, []byte(constitution))
}
