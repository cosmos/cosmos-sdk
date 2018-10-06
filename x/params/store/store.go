package store

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Additional capicity to be allocated for Store.space
// So we don't have to allocate extra space each time appending to the key
const extraKeyCap = 20

// Individual parameter store for each keeper
// Transient store persists for a block, so we use it for
// recording whether the parameter has been changed or not
type Store struct {
	cdc  *codec.Codec
	key  sdk.StoreKey // []byte -> []byte, stores parameter
	tkey sdk.StoreKey // []byte -> bool, stores parameter change

	space []byte

	table Table
}

// NewStore constructs a store with namestore
func NewStore(cdc *codec.Codec, key sdk.StoreKey, tkey sdk.StoreKey, space string, table Table) (res Store) {
	res = Store{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		table: table,
	}

	spacebz := []byte(space)
	res.space = make([]byte, len(spacebz), len(spacebz)+extraKeyCap)
	copy(res.space, spacebz)
	return
}

// Returns a KVStore identical with ctx.KVStore(s.key).Prefix()
func (s Store) kvStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return ctx.KVStore(s.key).Prefix(append(s.space, '/'))
}

// Returns a KVStore identical with ctx.TransientStore(s.tkey).Prefix()
func (s Store) transientStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return ctx.TransientStore(s.tkey).Prefix(append(s.space, '/'))
}

// Get parameter from store
func (s Store) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	store := s.kvStore(ctx)
	bz := store.Get(key)
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// GetIfExists do not modify ptr if the stored parameter is nil
func (s Store) GetIfExists(ctx sdk.Context, key []byte, ptr interface{}) {
	store := s.kvStore(ctx)
	bz := store.Get(key)
	if bz == nil {
		return
	}
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// Get raw bytes of parameter from store
func (s Store) GetRaw(ctx sdk.Context, key []byte) []byte {
	store := s.kvStore(ctx)
	return store.Get(key)
}

// Check if the parameter is set in the store
func (s Store) Has(ctx sdk.Context, key []byte) bool {
	store := s.kvStore(ctx)
	return store.Has(key)
}

// Returns true if the parameter is set in the block
func (s Store) Modified(ctx sdk.Context, key []byte) bool {
	tstore := s.transientStore(ctx)
	return tstore.Has(key)
}

// Set parameter, return error if stored parameter has different type from input
// Also set to the transient store to record change
func (s Store) Set(ctx sdk.Context, key []byte, param interface{}) {
	store := s.kvStore(ctx)

	ty, ok := s.table[string(key)]
	if !ok {
		panic("Parameter not registered")
	}

	pty := reflect.TypeOf(param)
	if pty.Kind() == reflect.Ptr {
		pty = pty.Elem()
	}

	if pty != ty {
		panic("Type mismatch with registered table")
	}

	bz, err := s.cdc.MarshalJSON(param)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)

	tstore := s.transientStore(ctx)
	tstore.Set(key, []byte{})
}

// Get to ParamStruct
func (s Store) GetStruct(ctx sdk.Context, ps ParamStruct) {
	for _, pair := range ps.KeyValuePairs() {
		s.Get(ctx, pair.Key, pair.Value)
	}
}

// Set from ParamStruct
func (s Store) SetStruct(ctx sdk.Context, ps ParamStruct) {
	for _, pair := range ps.KeyValuePairs() {
		// pair.Field is a pointer to the field, so indirecting the ptr.
		// go-amino automatically handles it but just for sure,
		// since SetStruct is meant to be used in InitGenesis
		// so this method will not be called frequently
		v := reflect.Indirect(reflect.ValueOf(pair.Value)).Interface()
		s.Set(ctx, pair.Key, v)
	}
}

// Returns internal namespace
func (s Store) Space() string {
	return string(s.space)
}

// Wrapper of Store, provides immutable functions only
type ReadOnlyStore struct {
	s Store
}

// Exposes Get
func (ros ReadOnlyStore) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlyStore) GetRaw(ctx sdk.Context, key []byte) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlyStore) Has(ctx sdk.Context, key []byte) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlyStore) Modified(ctx sdk.Context, key []byte) bool {
	return ros.s.Modified(ctx, key)
}

// Exposes Space
func (ros ReadOnlyStore) Space() string {
	return ros.s.Space()
}
