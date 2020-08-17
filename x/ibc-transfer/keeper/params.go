package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

// GetTransfersEnabled retrieves the transfers enabled boolean from the paramstore
func (k Keeper) GetTransfersEnabled(ctx sdk.Context) bool {
	var res bool
	k.paramSpace.Get(ctx, types.KeyTransfersEnabled, &res)
	return res
}

// GetParams returns the total set of ibc-transfer parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	return types.NewParams(k.GetTransfersEnabled(ctx))
}

// SetParams sets the total set of ibc-transfer parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
