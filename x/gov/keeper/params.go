package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// SetParams sets the gov module's parameters.
// CONTRACT: This method performs no validation of the parameters.
func (k Keeper) SetParams(ctx context.Context, params v1.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// GetParams gets the gov module's parameters.
func (k Keeper) GetParams(ctx context.Context) (params v1.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		return params, err
	}
	if bz == nil {
		return params, nil
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}
