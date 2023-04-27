package keeper

import (
	"context"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// GetParams returns the total set of distribution parameters.
func (k Keeper) GetParams(ctx context.Context) (params types.Params, err error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.ParamsKey)
	if bz == nil || err != nil {
		return params, err
	}

	err = k.cdc.Unmarshal(bz, &params)
	return params, err
}

// SetParams sets the distribution parameters.
// CONTRACT: This method performs no validation of the parameters.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return store.Set(types.ParamsKey, bz)
}

// GetCommunityTax returns the current distribution community tax.
func (k Keeper) GetCommunityTax(ctx context.Context) (math.LegacyDec, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return math.LegacyDec{}, err
	}

	return params.CommunityTax, nil
}

// GetWithdrawAddrEnabled returns the current distribution withdraw address
// enabled parameter.
func (k Keeper) GetWithdrawAddrEnabled(ctx context.Context) (enabled bool, err error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return false, err
	}

	return params.WithdrawAddrEnabled, nil
}
