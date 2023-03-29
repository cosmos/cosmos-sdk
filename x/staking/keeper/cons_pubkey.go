package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) SetConsPubKeyRotationHistory(ctx sdk.Context, valAddr sdk.ValAddress, oldPubKey, newPubKey *codectypes.Any, height uint64) {
	history := types.ConsPubKeyRotationHistory{
		OperatorAddress: valAddr.String(),
		OldConsPubkey:   oldPubKey,
		NewConsPubkey:   newPubKey,
		Height:          height,
	}
	store := ctx.KVStore(k.storeKey)
	key := types.GetValidatorConsPubKeyRotationHistoryKey(history)
	historyBytes := k.cdc.MustMarshal(&history)
	store.Set(key, historyBytes)

	key = types.GetBlockConsPubKeyRotationHistoryKey(history)
	store.Set(key, historyBytes)
	queueTime := ctx.BlockHeader().Time.Add(k.UnbondingTime(ctx))

	k.SetConsKeyQueue(ctx, queueTime, valAddr)
	k.SetConsKeyIndex(ctx, valAddr, queueTime)
}

func (k Keeper) GetValidatorConsPubKeyRotationHistory(ctx sdk.Context, operatorAddress sdk.ValAddress) (historyObjects []types.ConsPubKeyRotationHistory) {
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store, types.GetValidatorConsPubKeyRotationHistoryPrefix(operatorAddress.String()))
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

func (k Keeper) GetConsKeyQueue(ctx sdk.Context, ts time.Time) types.ValAddrsOfRotatedConsKeys {
	var valAddrs types.ValAddrsOfRotatedConsKeys
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsKeyRotationTimeKey(ts)
	bz := store.Get(key)
	if bz == nil {
		return valAddrs
	}
	k.cdc.MustUnmarshal(bz, &valAddrs)
	return valAddrs
}

func (k Keeper) SetConsKeyQueue(ctx sdk.Context, ts time.Time, valAddr sdk.ValAddress) {
	operKeys := k.GetConsKeyQueue(ctx, ts)
	operKeys.Addresses = append(operKeys.Addresses, valAddr.String())
	key := types.GetConsKeyRotationTimeKey(ts)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&operKeys)
	store.Set(key, bz)
}

func (k Keeper) SetConsKeyIndex(ctx sdk.Context, valAddr sdk.ValAddress, ts time.Time) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsKeyIndexKey(valAddr, ts)
	store.Set(key, []byte{})
}

func (k Keeper) UpdateAllMaturedConsKeyRotatedKeys(ctx sdk.Context, maturedTime time.Time) {
	maturedRotatedValAddrs := k.GetAllMaturedRotatedKeys(ctx, maturedTime)
	for _, valAddrStr := range maturedRotatedValAddrs {
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		if err != nil {
			panic(err)
		}

		k.deleteConsKeyIndexKey(ctx, valAddr, maturedTime)
	}
}

func (k Keeper) GetAllMaturedRotatedKeys(ctx sdk.Context, matureTime time.Time) []string {
	store := ctx.KVStore(k.storeKey)
	var ValAddresses []string
	iterator := store.Iterator(types.ValidatorConsensusKeyRotationRecordQueueKey, storetypes.InclusiveEndBytes(types.GetConsKeyRotationTimeKey(matureTime)))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var operKey types.ValAddrsOfRotatedConsKeys
		k.cdc.MustUnmarshal(iterator.Value(), &operKey)
		ValAddresses = append(ValAddresses, operKey.Addresses...)
		store.Delete(iterator.Key())
	}

	return ValAddresses
}

func (k Keeper) deleteConsKeyIndexKey(ctx sdk.Context, valAddr sdk.ValAddress, ts time.Time) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetConsKeyIndexKey(valAddr, ts)
	iterator := store.Iterator(types.ValidatorConsensusKeyRotationRecordIndexKey, storetypes.InclusiveEndBytes(key))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) CheckLimitOfMaxRotationsExceed(ctx sdk.Context, valAddr sdk.ValAddress) (bool, uint64) {
	store := ctx.KVStore(k.storeKey)
	key := append(types.ValidatorConsensusKeyRotationRecordIndexKey, address.MustLengthPrefix(valAddr)...)
	prefixIterator := storetypes.KVStorePrefixIterator(store, key)
	defer prefixIterator.Close()

	count := uint64(0)
	for ; prefixIterator.Valid(); prefixIterator.Next() {
		count += 1
	}

	return count >= k.MaxConsPubKeyRotations(ctx), count
}
