package keeper

import (
	"time"

	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetValidatorOperatorKeyRotationsHistory(ctx sdk.Context,
	valAddr sdk.ValAddress) []types.OperatorKeyRotationRecord {

	var historyObjs []types.OperatorKeyRotationRecord
	store := ctx.KVStore(k.storeKey)

	iterator := storetypes.KVStorePrefixIterator(store,
		append([]byte(types.ValidatorOperatorKeyRotationRecordKey), valAddr...))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var history types.OperatorKeyRotationRecord

		k.cdc.MustUnmarshal(iterator.Value(), &history)
		historyObjs = append(historyObjs, history)
	}
	return historyObjs
}

func (k Keeper) SetOperatorKeyRotationRecord(ctx sdk.Context, curValAddr, newValAddr sdk.ValAddress, height int64) error {
	historyObj := types.OperatorKeyRotationRecord{
		OperatorAddress:    newValAddr.String(),
		OldOperatorAddress: curValAddr.String(),
		Height:             uint64(ctx.BlockHeight()),
	}

	store := ctx.KVStore(k.storeKey)
	key := types.GetOperKeyRotationHistoryKey(newValAddr, uint64(height))
	bz, err := k.cdc.Marshal(&historyObj)

	if err != nil {
		return err
	}

	store.Set(key, bz)
	k.SetVORQueue(ctx, ctx.BlockTime(), newValAddr)
	k.SetVORIndex(ctx, newValAddr)

	return nil
}

func (k Keeper) GetVORQueue(ctx sdk.Context, ts time.Time) types.RotatedOperatorAddresses {
	var valAddrs types.RotatedOperatorAddresses
	store := ctx.KVStore(k.storeKey)
	key := types.GetOperatorRotationTimeKey(ts)
	bz := store.Get(key)
	if bz == nil {
		return valAddrs
	}
	k.cdc.MustUnmarshal(bz, &valAddrs)
	return valAddrs
}

func (k Keeper) SetVORQueue(ctx sdk.Context, ts time.Time, valAddr sdk.ValAddress) {
	operKeys := k.GetVORQueue(ctx, ts)
	operKeys.Addresses = append(operKeys.Addresses, valAddr.String())
	key := types.GetOperatorRotationTimeKey(ts)
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&operKeys)
	store.Set(key, bz)
}

func (k Keeper) SetVORIndex(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetVORIndexKey(valAddr)
	store.Set(key, []byte{})
}

func (k Keeper) GetVORIndex(ctx sdk.Context, valAddr sdk.ValAddress) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.GetVORIndexKey(valAddr)
	bz := store.Get(key)
	return bz != nil
}

func (k Keeper) UpdateAllMaturedVORRotatedKeys(ctx sdk.Context, maturedTime time.Time) {
	maturedRotatedValAddrs := k.GetAllMaturedRotatedKeys(ctx, maturedTime)
	for _, valAddrStr := range maturedRotatedValAddrs {
		valAddr, err := sdk.ValAddressFromBech32(valAddrStr)
		if err != nil {
			panic(err)
		}

		k.deleteVORIndexKey(ctx, valAddr)
	}
}

func (k Keeper) GetAllMaturedRotatedKeys(ctx sdk.Context, matureTime time.Time) []string {
	store := ctx.KVStore(k.storeKey)
	var ValAddresses []string
	prefixIterator := storetypes.KVStorePrefixIterator(store, storetypes.InclusiveEndBytes(types.GetOperatorRotationTimeKey(matureTime)))

	for ; prefixIterator.Valid(); prefixIterator.Next() {
		var operKey types.RotatedOperatorAddresses
		k.cdc.MustUnmarshal(prefixIterator.Value(), &operKey)
		ValAddresses = append(ValAddresses, operKey.Addresses...)
		store.Delete(prefixIterator.Key())
	}

	return ValAddresses
}

func (k Keeper) deleteVORIndexKey(ctx sdk.Context, valAddr sdk.ValAddress) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetVORIndexKey(valAddr)
	store.Delete(key)
}
