package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// SignedBlocksWindow - sliding window for downtime slashing
func (k Keeper) SignedBlocksWindow(ctx sdk.Context) (res int64) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeySignedBlocksWindow)
	if bz == nil {
		return res
	}
	k.legacyAmino.Unmarshal(bz, &res)
	return res
}

// MinSignedPerWindow - minimum blocks signed per window
func (k Keeper) MinSignedPerWindow(ctx sdk.Context) int64 {
	var minSignedPerWindow sdk.Dec

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyMinSignedPerWindow)
	if bz == nil {
		return minSignedPerWindow.RoundInt64()
	}
	k.legacyAmino.Unmarshal(bz, &minSignedPerWindow)
	signedBlocksWindow := k.SignedBlocksWindow(ctx)

	// NOTE: RoundInt64 will never panic as minSignedPerWindow is
	//       less than 1.
	return minSignedPerWindow.MulInt64(signedBlocksWindow).RoundInt64()
}

// DowntimeJailDuration - Downtime unbond duration
func (k Keeper) DowntimeJailDuration(ctx sdk.Context) (res time.Duration) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeyDowntimeJailDuration)
	if bz == nil {
		return res
	}
	k.legacyAmino.Unmarshal(bz, &res)
	return res
}

// SlashFractionDoubleSign - fraction of power slashed in case of double sign
func (k Keeper) SlashFractionDoubleSign(ctx sdk.Context) (res sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeySlashFractionDoubleSign)
	if bz == nil {
		return res
	}
	k.legacyAmino.Unmarshal(bz, &res)
	return res
}

// SlashFractionDowntime - fraction of power slashed for downtime
func (k Keeper) SlashFractionDowntime(ctx sdk.Context) (res sdk.Dec) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.KeySlashFractionDowntime)
	if bz == nil {
		return res
	}
	k.legacyAmino.Unmarshal(bz, &res)
	return res
}

// GetParams returns the current x/slashing module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return params
	}
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the x/slashing module parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}
