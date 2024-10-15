package keeper

import (
	"context"

	"cosmossdk.io/x/bank/v2/types"
)

// GetAuthorityMetadata returns the authority metadata for a specific denom
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.params.Set(ctx, params)
}

// setAuthorityMetadata stores authority metadata for a specific denom
func (k Keeper) GetParams(ctx context.Context) types.Params {
	params, err := k.params.Get(ctx)
	if err != nil {
		return types.Params{}
	}
	return params
}
