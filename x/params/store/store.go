package store

import (
	"fmt"
	"reflect"

	tmlibs "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Individual parameter store for each keeper
type Store struct {
	cdc  *codec.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	store []byte
}

// NewStore constructs a store with namestore
func NewStore(cdc *codec.Codec, key sdk.StoreKey, tkey sdk.StoreKey, store string) Store {
	if !tmlibs.IsASCIIText(store) {
		panic("paramstore store expressions can only contain alphanumeric characters")
	}

	return Store{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		store: append([]byte(store), '/'),
	}
}

// Get parameter from store
func (s Store) Get(ctx sdk.Context, key string, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.store)
	bz := store.Get([]byte(key))
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// GetIfExists do not modify ptr if the stored parameter is nil
func (s Store) GetIfExists(ctx sdk.Context, key string, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.store)
	bz := store.Get([]byte(key))
	if bz == nil {
		return
	}
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// Get raw bytes of parameter from store
func (s Store) GetRaw(ctx sdk.Context, key string) []byte {
	store := ctx.KVStore(s.key).Prefix(s.store)
	res := store.Get([]byte(key))
	return res
}

// Check if the parameter is set in the store
func (s Store) Has(ctx sdk.Context, key string) bool {
	store := ctx.KVStore(s.key).Prefix(s.store)
	return store.Has([]byte(key))
}

// Returns true if the parameter is set in the block
func (s Store) Modified(ctx sdk.Context, key string) bool {
	tstore := ctx.KVStore(s.tkey).Prefix(s.store)
	return tstore.Has([]byte(key))
}

// Set parameter, return error if stored parameter has different type from input
func (s Store) Set(ctx sdk.Context, key string, param interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.store)
	keybz := []byte(key)

	bz := store.Get(keybz)
	if bz != nil {
		ptrty := reflect.PtrTo(reflect.TypeOf(param))
		ptr := reflect.New(ptrty).Interface()

		if s.cdc.UnmarshalJSON(bz, ptr) != nil {
			panic(fmt.Errorf("Type mismatch with stored param and provided param"))
		}
	}

	bz, err := s.cdc.MarshalJSON(param)
	if err != nil {
		panic(err)
	}
	store.Set(keybz, bz)

	tstore := ctx.KVStore(s.tkey).Prefix(s.store)
	tstore.Set(keybz, []byte{})
}

// Set raw bytes of parameter
func (s Store) SetRaw(ctx sdk.Context, key string, param []byte) {
	keybz := []byte(key)

	store := ctx.KVStore(s.key).Prefix(s.store)
	store.Set(keybz, param)

	tstore := ctx.KVStore(s.tkey).Prefix(s.store)
	tstore.Set(keybz, []byte{})
}

// Set from ParamStruct
func (s Store) SetFromParamStruct(ctx sdk.Context, ps ParamStruct) {
	for _, pair := range ps.KeyFieldPairs() {
		s.Set(ctx, pair.Key, pair.Field)
	}
}

// Returns a KVStore identical with the paramstore
func (s Store) KVStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(s.key).Prefix(s.store)
}

// Wrapper of Store, provides immutable functions only
type ReadOnlyStore struct {
	s Store
}

// Exposes Get
func (ros ReadOnlyStore) Get(ctx sdk.Context, key string, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlyStore) GetRaw(ctx sdk.Context, key string) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlyStore) Has(ctx sdk.Context, key string) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlyStore) Modified(ctx sdk.Context, key string) bool {
	return ros.s.Modified(ctx, key)
}
