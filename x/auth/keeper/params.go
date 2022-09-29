package keeper

import (
	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// SetParams sets the auth module's parameters.
func (ak AccountKeeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(ak.storeKey)
	bz := ak.cdc.MustMarshal(&params)
	store2.Set(store, types.ParamsKey, bz)

	return nil
}

// GetParams gets the auth module's parameters.
func (ak AccountKeeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(ak.storeKey)
	params, err := store2.GetAndDecode(store, ak.decodeParams, types.ParamsKey)
	if err != nil {
		panic(err)
	}
	return params
}
