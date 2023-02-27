package keeper

import (
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) SetConsPubKeyRotationHistory(ctx sdk.Context, history types.ConsPubKeyRotationHistory) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorConsPubKeyRotationHistoryKey(history)
	historyBytes := k.cdc.MustMarshal(&history)
	store.Set(key, historyBytes)

	key = types.GetBlockConsPubKeyRotationHistoryKey(history)
	store.Set(key, historyBytes)
}

func (k Keeper) GetValidatorConsPubKeyRotationHistory(ctx sdk.Context, operatorAddress sdk.ValAddress) (historyObjects []types.ConsPubKeyRotationHistory) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.GetValdiatorConsPubKeyRotationHistoryPrefix(operatorAddress.String()))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var history types.ConsPubKeyRotationHistory

		k.cdc.MustUnmarshal(iterator.Value(), &history)
		historyObjects = append(historyObjects, history)
	}
	return
}

func (k Keeper) GetBlockConsPubKeyRotationHistory(ctx sdk.Context, height int64) (historyObjects []types.ConsPubKeyRotationHistory) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.GetBlockConsPubKeyRotationHistoryPrefix(height))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var history types.ConsPubKeyRotationHistory

		k.cdc.MustUnmarshal(iterator.Value(), &history)
		historyObjects = append(historyObjects, history)
	}
	return
}
