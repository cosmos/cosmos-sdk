package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
)

// GetAllowedClients retrieves the list of allowed client types.
func (k Keeper) GetAllowedClients(ctx sdk.Context) []string {
	var res []string
	k.paramSpace.Get(ctx, types.KeyAllowedClients, &res)
	return res
}

// HistoricalEntries retrieves the number of historical info entries
// to persist in store
func (k Keeper) HistoricalEntries(ctx sdk.Context) uint32 {
	var res uint32
	k.paramSpace.Get(ctx, types.KeyHistoricalEntries, &res)
	return res
}

// GetParams returns the total set of ibc client parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.HistoricalEntries(ctx), k.GetAllowedClients(ctx)...)
}

// SetParams sets the total set of ibc client parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}
