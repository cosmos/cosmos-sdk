package keeper

import (
	"fmt"

	gogotypes "github.com/gogo/protobuf/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetLastTokenizeShareRecordID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bytes := store.Get(types.LastTokenizeShareRecordIDKey)
	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

func (k Keeper) SetLastTokenizeShareRecordID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.LastTokenizeShareRecordIDKey, sdk.Uint64ToBigEndian(id))
}

func (k Keeper) GetTokenizeShareRecord(ctx sdk.Context, id uint64) (tokenizeShareRecord types.TokenizeShareRecord, err error) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetTokenizeShareRecordByIndexKey(id))
	if bz == nil {
		return tokenizeShareRecord, sdkerrors.Wrap(types.ErrTokenizeShareRecordNotExists, fmt.Sprintf("tokenizeShareRecord %d does not exist", id))
	}

	k.cdc.MustUnmarshal(bz, &tokenizeShareRecord)
	return tokenizeShareRecord, nil
}

func (k Keeper) GetTokenizeShareRecordsByOwner(ctx sdk.Context, owner sdk.AccAddress) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)

	it := sdk.KVStorePrefixIterator(store, types.GetTokenizeShareRecordIdsByOwnerPrefix(owner))
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var id gogotypes.UInt64Value
		k.cdc.MustUnmarshal(it.Value(), &id)

		tokenizeShareRecord, err := k.GetTokenizeShareRecord(ctx, id.Value)
		if err != nil {
			continue
		}
		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) GetTokenizeShareRecordByDenom(ctx sdk.Context, denom string) (types.TokenizeShareRecord, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetTokenizeShareRecordIDByDenomKey(denom))
	if bz == nil {
		return types.TokenizeShareRecord{}, fmt.Errorf("tokenize share record not found from denom: %s", denom)
	}

	var id gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &id)

	return k.GetTokenizeShareRecord(ctx, id.Value)
}

func (k Keeper) GetAllTokenizeShareRecords(ctx sdk.Context) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)

	it := sdk.KVStorePrefixIterator(store, types.TokenizeShareRecordPrefix)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var tokenizeShareRecord types.TokenizeShareRecord
		k.cdc.MustUnmarshal(it.Value(), &tokenizeShareRecord)

		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) AddTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) error {
	if k.hasTokenizeShareRecord(ctx, tokenizeShareRecord.Id) {
		return sdkerrors.Wrapf(types.ErrTokenizeShareRecordAlreadyExists, "TokenizeShareRecord already exists: %d", tokenizeShareRecord.Id)
	}

	k.setTokenizeShareRecord(ctx, tokenizeShareRecord)

	owner, err := sdk.AccAddressFromBech32(tokenizeShareRecord.Owner)
	if err != nil {
		return err
	}

	k.setTokenizeShareRecordWithOwner(ctx, owner, tokenizeShareRecord.Id)
	k.setTokenizeShareRecordWithDenom(ctx, tokenizeShareRecord.GetShareTokenDenom(), tokenizeShareRecord.Id)

	return nil
}

func (k Keeper) DeleteTokenizeShareRecord(ctx sdk.Context, recordID uint64) error {
	record, err := k.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}
	owner, err := sdk.AccAddressFromBech32(record.Owner)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTokenizeShareRecordByIndexKey(recordID))
	store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, recordID))
	store.Delete(types.GetTokenizeShareRecordIDByDenomKey(record.GetShareTokenDenom()))
	return nil
}

func (k Keeper) hasTokenizeShareRecord(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.GetTokenizeShareRecordByIndexKey(id))
}

func (k Keeper) setTokenizeShareRecord(ctx sdk.Context, tokenizeShareRecord types.TokenizeShareRecord) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&tokenizeShareRecord)

	store.Set(types.GetTokenizeShareRecordByIndexKey(tokenizeShareRecord.Id), bz)
}

func (k Keeper) setTokenizeShareRecordWithOwner(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	store.Set(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id), bz)
}

func (k Keeper) deleteTokenizeShareRecordWithOwner(ctx sdk.Context, owner sdk.AccAddress, id uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id))
}

func (k Keeper) setTokenizeShareRecordWithDenom(ctx sdk.Context, denom string, id uint64) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	store.Set(types.GetTokenizeShareRecordIDByDenomKey(denom), bz)
}
