package store

import (
	"fmt"
	"reflect"
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Individual parameter store for each keeper
type Store struct {
	cdc  *wire.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	space string
}

// NewStore constructs a store with namespace
func NewStore(cdc *wire.Codec, key sdk.StoreKey, tkey sdk.StoreKey, space string) Store {
	return Store{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		space: space,
	}
}

// Wrapper for key string
type Key struct {
	s string
}

// copied from baseapp/router.go
var isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString

// Appending two keys with '/' as separator
// Checks alpanumericity
func (k Key) Append(keys ...string) (res Key) {
	res = k

	for _, key := range keys {
		if !isAlphaNumeric(key) {
			panic("parameter key expressions can only contain alphanumeric characters")
		}
		res.s = res.s + "/" + key
	}
	return
}

// NewKey constructs a key from a list of strings
func NewKey(keys ...string) (res Key) {
	res = Key{""}

	return res.Append(keys...)
}

// KeyBytes make KVStore key bytes from Key
func (k Key) KeyBytes(space string) []byte {
	return append([]byte(space), []byte(k.s)...)
}

// Human readable string
func (k Key) String() string {
	return k.s
}

// Get parameter from store
func (s Store) Get(ctx sdk.Context, key Key, ptr interface{}) {
	store := ctx.KVStore(s.key)
	bz := store.Get(key.KeyBytes(s.space))
	s.cdc.MustUnmarshalBinary(bz, ptr)
}

// Get raw bytes of parameter from store
func (s Store) GetRaw(ctx sdk.Context, key Key) []byte {
	store := ctx.KVStore(s.key)
	return store.Get(key.KeyBytes(s.space))
}

// Check if the parameter is set in the store
func (s Store) Has(ctx sdk.Context, key Key) bool {
	store := ctx.KVStore(s.key)
	return store.Has(key.KeyBytes(s.space))
}

// Returns true if the parameter is set in the block
func (s Store) Modified(ctx sdk.Context, key Key) bool {
	tstore := ctx.KVStore(s.tkey)
	return tstore.Has(key.KeyBytes(s.space))
}

// Set parameter, return error if stored parameter has different type from input
func (s Store) Set(ctx sdk.Context, key Key, param interface{}) error {
	store := ctx.KVStore(s.key)
	keybz := key.KeyBytes(s.space)
	bz := store.Get(keybz)
	if bz != nil {
		ptrty := reflect.PtrTo(reflect.TypeOf(param))
		ptr := reflect.New(ptrty).Interface()

		if s.cdc.UnmarshalBinary(bz, ptr) != nil {
			return fmt.Errorf("Type mismatch with stored param and provided param")
		}
	}

	bz, err := s.cdc.MarshalBinary(param)
	if err != nil {
		return err
	}
	store.Set(keybz, bz)

	tstore := ctx.KVStore(s.tkey)
	tstore.Set(keybz, []byte{})

	return nil
}

// Set raw bytes of parameter
func (s Store) SetRaw(ctx sdk.Context, key Key, param []byte) {
	keybz := key.KeyBytes(s.space)

	store := ctx.KVStore(s.key)
	store.Set(keybz, param)

	tstore := ctx.KVStore(s.tkey)
	tstore.Set(keybz, []byte{})
}

// Wrapper of Store, provides only immutable functions
type ReadOnlyStore struct {
	s Store
}

// Exposes Get
func (ros ReadOnlyStore) Get(ctx sdk.Context, key Key, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlyStore) GetRaw(ctx sdk.Context, key Key) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlyStore) Has(ctx sdk.Context, key Key) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlyStore) Modified(ctx sdk.Context, key Key) bool {
	return ros.s.Modified(ctx, key)
}
