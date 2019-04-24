package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// ParamKeyTable for minting module
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyParams, types.Params{},
	)
}

// get inflation params from the global param store
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramSpace.Get(ctx, ParamStoreKeyParams, &params)
	return params
}

// set inflation params from the global param store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.Set(ctx, ParamStoreKeyParams, &params)
}
