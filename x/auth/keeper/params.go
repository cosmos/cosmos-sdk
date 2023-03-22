package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SetParams sets the auth module's parameters.
func (ak AccountKeeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ak.storeSvc.OpenKVStore(ctx)
	bz := ak.cdc.MustMarshal(&params)
	return store.Set(types.ParamsKey, bz)
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ak.storeSvc.OpenKVStore(ctx)
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
