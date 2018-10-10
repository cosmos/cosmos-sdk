package subspace

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Additional capicity to be allocated for Subspace.name
// So we don't have to allocate extra space each time appending to the key
const extraKeyCap = 20

// Individual parameter store for each keeper
// Transient store persists for a block, so we use it for
// recording whether the parameter has been changed or not
type Subspace struct {
	cdc  *codec.Codec
	key  sdk.StoreKey // []byte -> []byte, stores parameter
	tkey sdk.StoreKey // []byte -> bool, stores parameter change

	name []byte

	table TypeTable
}

// NewSubspace constructs a store with namestore
func NewSubspace(cdc *codec.Codec, key sdk.StoreKey, tkey sdk.StoreKey, name string) (res Subspace) {
	res = Subspace{
		cdc:  cdc,
		key:  key,
		tkey: tkey,
	}

	namebz := []byte(name)
	res.name = make([]byte, len(namebz), len(namebz)+extraKeyCap)
	copy(res.name, namebz)
	return
}

// WithTypeTable initializes TypeTable and returns modified Subspace
func (s Subspace) WithTypeTable(table TypeTable) (res Subspace) {
	if table == nil {
		panic("SetTypeTable() called with nil TypeTable")
	}
	if s.table != nil {
		panic("SetTypeTable() called on initialized Subspace")
	}

	res = Subspace{
		cdc:   s.cdc,
		key:   s.key,
		tkey:  s.tkey,
		name:  s.name,
		table: table,
	}

	return
}

// Returns a KVStore identical with ctx.KVStore(s.key).Prefix()
func (s Subspace) kvStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return ctx.KVStore(s.key).Prefix(append(s.name, '/'))
}

// Returns a KVStore identical with ctx.TransientStore(s.tkey).Prefix()
func (s Subspace) transientStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return ctx.TransientStore(s.tkey).Prefix(append(s.name, '/'))
}

// Get parameter from store
func (s Subspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	store := s.kvStore(ctx)
	bz := store.Get(key)
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// GetIfExists do not modify ptr if the stored parameter is nil
func (s Subspace) GetIfExists(ctx sdk.Context, key []byte, ptr interface{}) {
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
func (s Subspace) GetRaw(ctx sdk.Context, key []byte) []byte {
	store := s.kvStore(ctx)
	return store.Get(key)
}

// Check if the parameter is set in the store
func (s Subspace) Has(ctx sdk.Context, key []byte) bool {
	store := s.kvStore(ctx)
	return store.Has(key)
}

// Returns true if the parameter is set in the block
func (s Subspace) Modified(ctx sdk.Context, key []byte) bool {
	tstore := s.transientStore(ctx)
	return tstore.Has(key)
}

// Set parameter, return error if stored parameter has different type from input
// Also set to the transient store to record change
func (s Subspace) Set(ctx sdk.Context, key []byte, param interface{}) {
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

// Get to ParamSet
func (s Subspace) GetParamSet(ctx sdk.Context, ps ParamSet) {
	for _, pair := range ps.KeyValuePairs() {
		s.Get(ctx, pair.Key, pair.Value)
	}
}

// Set from ParamSet
func (s Subspace) SetParamSet(ctx sdk.Context, ps ParamSet) {
	for _, pair := range ps.KeyValuePairs() {
		// pair.Field is a pointer to the field, so indirecting the ptr.
		// go-amino automatically handles it but just for sure,
		// since SetStruct is meant to be used in InitGenesis
		// so this method will not be called frequently
		v := reflect.Indirect(reflect.ValueOf(pair.Value)).Interface()
		s.Set(ctx, pair.Key, v)
	}
}

// Returns name of Subspace
func (s Subspace) Name() string {
	return string(s.name)
}

// Wrapper of Subspace, provides immutable functions only
type ReadOnlySubspace struct {
	s Subspace
}

// Exposes Get
func (ros ReadOnlySubspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlySubspace) GetRaw(ctx sdk.Context, key []byte) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlySubspace) Has(ctx sdk.Context, key []byte) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlySubspace) Modified(ctx sdk.Context, key []byte) bool {
	return ros.s.Modified(ctx, key)
}

// Exposes Space
func (ros ReadOnlySubspace) Name() string {
	return ros.s.Name()
}
