package keeper

import (
	"cosmossdk.io/math"

	store2 "github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func (k Keeper) decodeParams(bz []byte) (types.Params, error) {
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(clientCtx sdk.Context) (params types.Params) {
	store := clientCtx.KVStore(k.storeKey)
	params, _ = store2.GetAndDecode(store, k.decodeParams, types.ParamsKey)
	return params
}

// SetParams sets the distribution parameters to the param space.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.ValidateBasic(); err != nil {
		return err
	}

	store := k.getStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetCommunityTax returns the current distribution community tax.
func (k Keeper) GetCommunityTax(ctx sdk.Context) math.LegacyDec {
	return k.GetParams(ctx).CommunityTax
}

// GetWithdrawAddrEnabled returns the current distribution withdraw address
// enabled parameter.
func (k Keeper) GetWithdrawAddrEnabled(ctx sdk.Context) (enabled bool) {
	return k.GetParams(ctx).WithdrawAddrEnabled
}
