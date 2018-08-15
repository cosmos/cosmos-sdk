package space

import (
	"fmt"
	"reflect"

	tmlibs "github.com/tendermint/tendermint/libs/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Individual parameter store for each keeper
type Space struct {
	cdc  *wire.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	space []byte
}

// NewSpace constructs a store with namespace
func NewSpace(cdc *wire.Codec, key sdk.StoreKey, tkey sdk.StoreKey, space string) Space {
	if !tmlibs.IsASCIIText(space) {
		panic("paramstore space expressions can only contain alphanumeric characters")
	}

	return Space{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		space: []byte(space + "/"),
	}
}

// Wrapper for key string
type Key struct {
	s string
}

// Appending two keys with '/' as separator
// Checks alpanumericity
func (k Key) Append(keys ...string) (res Key) {
	res = k

	for _, key := range keys {
		if !tmlibs.IsASCIIText(key) {
			panic("parameter key expressions can only contain alphanumeric characters")
		}
		res.s = res.s + "/" + key
	}
	return
}

// NewKey constructs a key from a list of strings
func NewKey(keys ...string) (res Key) {
	if len(keys) < 1 {
		panic("length of parameter keys must not be zero")
	}
	res = Key{keys[0]}

	return res.Append(keys[1:]...)
}

// KeyBytes make KVStore key bytes from Key
func (k Key) Bytes() []byte {
	return []byte(k.s)
}

// Human readable string
func (k Key) String() string {
	return k.s
}

// Get parameter from store
func (s Space) Get(ctx sdk.Context, key Key, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.space)
	bz := store.Get(key.Bytes())
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// GetIfExists do not modify ptr if the stored parameter is nil
func (s Space) GetIfExists(ctx sdk.Context, key Key, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.space)
	bz := store.Get(key.Bytes())
	if bz == nil {
		return
	}
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// Get raw bytes of parameter from store
func (s Space) GetRaw(ctx sdk.Context, key Key) []byte {
	store := ctx.KVStore(s.key).Prefix(s.space)
	res := store.Get(key.Bytes())
	return res
}

// Check if the parameter is set in the store
func (s Space) Has(ctx sdk.Context, key Key) bool {
	store := ctx.KVStore(s.key).Prefix(s.space)
	return store.Has(key.Bytes())
}

// Returns true if the parameter is set in the block
func (s Space) Modified(ctx sdk.Context, key Key) bool {
	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	return tstore.Has(key.Bytes())
}

// Set parameter, return error if stored parameter has different type from input
func (s Space) Set(ctx sdk.Context, key Key, param interface{}) error {
	store := ctx.KVStore(s.key).Prefix(s.space)
	keybz := key.Bytes()

	bz := store.Get(keybz)
	if bz != nil {
		ptrty := reflect.PtrTo(reflect.TypeOf(param))
		ptr := reflect.New(ptrty).Interface()

		if s.cdc.UnmarshalJSON(bz, ptr) != nil {
			return fmt.Errorf("Type mismatch with stored param and provided param")
		}
	}

	bz, err := s.cdc.MarshalJSON(param)
	if err != nil {
		return err
	}
	store.Set(keybz, bz)

	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	tstore.Set(keybz, []byte{})

	return nil
}

// Set raw bytes of parameter
func (s Space) SetRaw(ctx sdk.Context, key Key, param []byte) {
	keybz := key.Bytes()

	store := ctx.KVStore(s.key).Prefix(s.space)
	store.Set(keybz, param)

	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	tstore.Set(keybz, []byte{})
}

// Iterates over raw parameters in the substore

// Wrapper of Space, provides immutable functions only
type ReadOnlySpace struct {
	s Space
}

// Exposes Get
func (ros ReadOnlySpace) Get(ctx sdk.Context, key Key, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlySpace) GetRaw(ctx sdk.Context, key Key) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlySpace) Has(ctx sdk.Context, key Key) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlySpace) Modified(ctx sdk.Context, key Key) bool {
	return ros.s.Modified(ctx, key)
}
