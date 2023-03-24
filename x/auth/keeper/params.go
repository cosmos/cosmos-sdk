package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SetParams sets the auth module's parameters.
func (ak AccountKeeper) SetParams(ctx context.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ak.storeService.OpenKVStore(ctx)
	bz := ak.cdc.MustMarshal(&params)
	return store.Set(types.ParamsKey, bz)
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx context.Context) (params types.Params) {
	store := ak.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if err != nil {
		panic(err)
	}

	if bz == nil {
		return params
	}
	ak.cdc.MustUnmarshal(bz, &params)
	return params
}
