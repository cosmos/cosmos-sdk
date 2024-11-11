package keeper

import (
	"context"
	"fmt"

	gogotypes "github.com/cosmos/gogoproto/types"

	errorsmod "cosmossdk.io/errors"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (k Keeper) GetLastTokenizeShareRecordID(ctx context.Context) uint64 {
	store := k.storeService.OpenKVStore(ctx)
	bytes, err := store.Get(types.LastTokenizeShareRecordIDKey)
	if err != nil {
		panic(err)
	}

	if bytes == nil {
		return 0
	}
	return sdk.BigEndianToUint64(bytes)
}

func (k Keeper) SetLastTokenizeShareRecordID(ctx context.Context, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	err := store.Set(types.LastTokenizeShareRecordIDKey, sdk.Uint64ToBigEndian(id))
	if err != nil {
		panic(err)
	}
}

func (k Keeper) GetTokenizeShareRecord(ctx context.Context, id uint64) (tokenizeShareRecord types.TokenizeShareRecord, err error) {
	store := k.storeService.OpenKVStore(ctx)

	bz, err := store.Get(types.GetTokenizeShareRecordByIndexKey(id))
	if err != nil {
		return tokenizeShareRecord, err
	}

	if bz == nil {
		return tokenizeShareRecord, errorsmod.Wrap(types.ErrTokenizeShareRecordNotExists, fmt.Sprintf("tokenizeShareRecord %d does not exist", id))
	}

	k.cdc.MustUnmarshal(bz, &tokenizeShareRecord)
	return tokenizeShareRecord, nil
}

func (k Keeper) GetTokenizeShareRecordsByOwner(ctx context.Context, owner sdk.AccAddress) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := k.storeService.OpenKVStore(ctx)

	it := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.GetTokenizeShareRecordIdsByOwnerPrefix(owner))
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

func (k Keeper) GetTokenizeShareRecordByDenom(ctx context.Context, denom string) (types.TokenizeShareRecord, error) {
	store := k.storeService.OpenKVStore(ctx)
	bz, err := store.Get(types.GetTokenizeShareRecordIDByDenomKey(denom))
	if err != nil {
		return types.TokenizeShareRecord{}, err
	}

	if bz == nil {
		return types.TokenizeShareRecord{}, fmt.Errorf("tokenize share record not found from denom: %s", denom)
	}

	var id gogotypes.UInt64Value
	k.cdc.MustUnmarshal(bz, &id)

	return k.GetTokenizeShareRecord(ctx, id.Value)
}

func (k Keeper) GetAllTokenizeShareRecords(ctx context.Context) (tokenizeShareRecords []types.TokenizeShareRecord) {
	store := k.storeService.OpenKVStore(ctx)

	it := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.TokenizeShareRecordPrefix)
	defer it.Close()

	for ; it.Valid(); it.Next() {
		var tokenizeShareRecord types.TokenizeShareRecord
		k.cdc.MustUnmarshal(it.Value(), &tokenizeShareRecord)

		tokenizeShareRecords = append(tokenizeShareRecords, tokenizeShareRecord)
	}
	return
}

func (k Keeper) AddTokenizeShareRecord(ctx context.Context, tokenizeShareRecord types.TokenizeShareRecord) error {
	hasRecord, err := k.hasTokenizeShareRecord(ctx, tokenizeShareRecord.Id)
	if err != nil {
		return err
	}

	if hasRecord {
		return errorsmod.Wrapf(types.ErrTokenizeShareRecordAlreadyExists, "TokenizeShareRecord already exists: %d", tokenizeShareRecord.Id)
	}

	k.setTokenizeShareRecord(ctx, tokenizeShareRecord)

	owner, err := k.authKeeper.AddressCodec().StringToBytes(tokenizeShareRecord.Owner)
	if err != nil {
		return err
	}

	k.setTokenizeShareRecordWithOwner(ctx, owner, tokenizeShareRecord.Id)
	k.setTokenizeShareRecordWithDenom(ctx, tokenizeShareRecord.GetShareTokenDenom(), tokenizeShareRecord.Id)

	return nil
}

func (k Keeper) DeleteTokenizeShareRecord(ctx context.Context, recordID uint64) error {
	record, err := k.GetTokenizeShareRecord(ctx, recordID)
	if err != nil {
		return err
	}
	owner, err := k.authKeeper.AddressCodec().StringToBytes(record.Owner)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)
	err = store.Delete(types.GetTokenizeShareRecordByIndexKey(recordID))
	if err != nil {
		return err
	}
	err = store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, recordID))
	if err != nil {
		return err
	}
	err = store.Delete(types.GetTokenizeShareRecordIDByDenomKey(record.GetShareTokenDenom()))
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) hasTokenizeShareRecord(ctx context.Context, id uint64) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	return store.Has(types.GetTokenizeShareRecordByIndexKey(id))
}

func (k Keeper) setTokenizeShareRecord(ctx context.Context, tokenizeShareRecord types.TokenizeShareRecord) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&tokenizeShareRecord)

	err := store.Set(types.GetTokenizeShareRecordByIndexKey(tokenizeShareRecord.Id), bz)
	if err != nil {
		panic(err)
	}
}

func (k Keeper) setTokenizeShareRecordWithOwner(ctx context.Context, owner sdk.AccAddress, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	err := store.Set(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id), bz)
	if err != nil {
		panic(err)
	}
}

func (k Keeper) deleteTokenizeShareRecordWithOwner(ctx context.Context, owner sdk.AccAddress, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	err := store.Delete(types.GetTokenizeShareRecordIDByOwnerAndIDKey(owner, id))
	if err != nil {
		panic(err)
	}
}

func (k Keeper) setTokenizeShareRecordWithDenom(ctx context.Context, denom string, id uint64) {
	store := k.storeService.OpenKVStore(ctx)
	bz := k.cdc.MustMarshal(&gogotypes.UInt64Value{Value: id})

	err := store.Set(types.GetTokenizeShareRecordIDByDenomKey(denom), bz)
	if err != nil {
		panic(err)
	}
}
