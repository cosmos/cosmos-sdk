package keeper

import (
	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (k Keeper) SetParams(ctx sdk.Context, params v1.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store2.Set(store, types.ParamsKey, bz)

	return nil
}

func (k Keeper) decodeParams(bz []byte) (v1.Params, error) {
	var params v1.Params
	if bz == nil {
		return params, nil
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

func (k Keeper) GetParams(clientCtx sdk.Context) (params v1.Params) {
	store := clientCtx.KVStore(k.storeKey)
	params, err := store2.GetAndDecode(store, k.decodeParams, types.ParamsKey)
	if err != nil {
		panic(err)
	}
	return params
}
