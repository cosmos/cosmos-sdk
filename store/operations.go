package store

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	ErrStoreKeyNotFound = errors.New("key not found")
)

// Delete deletes the value in the store identified by sk and at the key
// identified by key. Delete will return an error if the key does not exist.
func Delete(ctx sdk.Context, sk StoreKey, key []byte) error {
	if !Has(ctx, sk, key) {
		return ErrStoreKeyNotFound
	}
	store := ctx.KVStore(sk)
	store.Delete(key)
	return nil
}

// Has returns true if the specified key exists in the store identified by sk.
func Has(ctx sdk.Context, sk StoreKey, key []byte) bool {
	store := ctx.KVStore(sk)
	return store.Has(key)
}

// IncrementSeq increments the Uint in the store identified by sk at the key seqKey.
func IncrementSeq(ctx sdk.Context, sk StoreKey, seqKey []byte) sdk.Uint {
	store := ctx.KVStore(sk)
	seq := GetSeq(ctx, sk, seqKey).Add(sdk.OneUint())
	store.Set(seqKey, []byte(seq.String()))
	return seq
}

// GetSeq returns the Uint in the store  identified by sk at the key seqKey.
func GetSeq(ctx sdk.Context, sk StoreKey, seqKey []byte) sdk.Uint {
	store := ctx.KVStore(sk)
	if !store.Has(seqKey) {
		return sdk.ZeroUint()
	}

	b := store.Get(seqKey)
	return sdk.NewUintFromString(string(b))
}
