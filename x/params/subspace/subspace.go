package subspace

import (
	"errors"
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
)

const (
	// StoreKey is the string store key for the param store
	StoreKey = "params"

	// TStoreKey is the string store key for the param transient store
	TStoreKey = "transient_params"
)

// Individual parameter store for each keeper
// Transient store persists for a block, so we use it for
// recording whether the parameter has been changed or not
type Subspace struct {
	cdc  *codec.Codec
	key  sdk.StoreKey // []byte -> []byte, stores parameter
	tkey sdk.StoreKey // []byte -> bool, stores parameter change

	name []byte

	table KeyTable
}

// NewSubspace constructs a store with namestore
func NewSubspace(cdc *codec.Codec, key sdk.StoreKey, tkey sdk.StoreKey, name string) (res Subspace) {
	res = Subspace{
		cdc:  cdc,
		key:  key,
		tkey: tkey,
		name: []byte(name),
		table: KeyTable{
			m: make(map[string]attribute),
		},
	}

	return
}

// WithKeyTable initializes KeyTable and returns modified Subspace
func (s Subspace) WithKeyTable(table KeyTable) Subspace {
	if table.m == nil {
		panic("SetKeyTable() called with nil KeyTable")
	}
	if len(s.table.m) != 0 {
		panic("SetKeyTable() called on already initialized Subspace")
	}

	for k, v := range table.m {
		s.table.m[k] = v
	}

	// Allocate additional capicity for Subspace.name
	// So we don't have to allocate extra space each time appending to the key
	name := s.name
	s.name = make([]byte, len(name), len(name)+table.maxKeyLength())
	copy(s.name, name)

	return s
}

// Returns a KVStore identical with ctx.KVStore(s.key).Prefix()
func (s Subspace) kvStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return prefix.NewStore(ctx.KVStore(s.key), append(s.name, '/'))
}

// Returns a transient store for modification
func (s Subspace) transientStore(ctx sdk.Context) sdk.KVStore {
	// append here is safe, appends within a function won't cause
	// weird side effects when its singlethreaded
	return prefix.NewStore(ctx.TransientStore(s.tkey), append(s.name, '/'))
}

func concatKeys(key, subkey []byte) (res []byte) {
	res = make([]byte, len(key)+1+len(subkey))
	copy(res, key)
	res[len(key)] = '/'
	copy(res[len(key)+1:], subkey)
	return
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

// GetWithSubkey returns a parameter with a given key and a subkey.
func (s Subspace) GetWithSubkey(ctx sdk.Context, key, subkey []byte, ptr interface{}) {
	s.Get(ctx, concatKeys(key, subkey), ptr)
}

// GetWithSubkeyIfExists  returns a parameter with a given key and a subkey but does not
// modify ptr if the stored parameter is nil.
func (s Subspace) GetWithSubkeyIfExists(ctx sdk.Context, key, subkey []byte, ptr interface{}) {
	s.GetIfExists(ctx, concatKeys(key, subkey), ptr)
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

func (s Subspace) checkType(store sdk.KVStore, key []byte, param interface{}) {
	attr, ok := s.table.m[string(key)]
	if !ok {
		panic("Parameter not registered")
	}

	ty := attr.ty
	pty := reflect.TypeOf(param)
	if pty.Kind() == reflect.Ptr {
		pty = pty.Elem()
	}

	if pty != ty {
		panic("Type mismatch with registered table")
	}
}

// Set stores the parameter. It returns error if stored parameter has different type from input.
// It also set to the transient store to record change.
func (s Subspace) Set(ctx sdk.Context, key []byte, param interface{}) {
	store := s.kvStore(ctx)

	s.checkType(store, key, param)

	bz, err := s.cdc.MarshalJSON(param)
	if err != nil {
		panic(err)
	}
	store.Set(key, bz)

	tstore := s.transientStore(ctx)
	tstore.Set(key, []byte{})

}

// Update stores raw parameter bytes. It returns error if the stored parameter
// has a different type from the input. It also sets to the transient store to
// record change.
func (s Subspace) Update(ctx sdk.Context, key []byte, param []byte) error {
	attr, ok := s.table.m[string(key)]
	if !ok {
		panic("Parameter not registered")
	}

	ty := attr.ty
	dest := reflect.New(ty).Interface()
	s.GetIfExists(ctx, key, dest)
	err := s.cdc.UnmarshalJSON(param, dest)
	if err != nil {
		return err
	}

	s.Set(ctx, key, dest)
	tStore := s.transientStore(ctx)
	tStore.Set(key, []byte{})

	return nil
}

// SetWithSubkey set a parameter with a key and subkey
// Checks parameter type only over the key
func (s Subspace) SetWithSubkey(ctx sdk.Context, key []byte, subkey []byte, param interface{}) {
	store := s.kvStore(ctx)

	s.checkType(store, key, param)

	newkey := concatKeys(key, subkey)

	bz, err := s.cdc.MarshalJSON(param)
	if err != nil {
		panic(err)
	}
	store.Set(newkey, bz)

	tstore := s.transientStore(ctx)
	tstore.Set(newkey, []byte{})
}

// UpdateWithSubkey stores raw parameter bytes  with a key and subkey. It checks
// the parameter type only over the key.
func (s Subspace) UpdateWithSubkey(ctx sdk.Context, key []byte, subkey []byte, param []byte) error {
	concatkey := concatKeys(key, subkey)

	attr, ok := s.table.m[string(concatkey)]
	if !ok {
		return errors.New("parameter not registered")
	}

	ty := attr.ty
	dest := reflect.New(ty).Interface()
	s.GetWithSubkeyIfExists(ctx, key, subkey, dest)
	err := s.cdc.UnmarshalJSON(param, dest)
	if err != nil {
		return err
	}

	s.SetWithSubkey(ctx, key, subkey, dest)
	tStore := s.transientStore(ctx)
	tStore.Set(concatkey, []byte{})

	return nil
}

// Get to ParamSet
func (s Subspace) GetParamSet(ctx sdk.Context, ps ParamSet) {
	for _, pair := range ps.ParamSetPairs() {
		s.Get(ctx, pair.Key, pair.Value)
	}
}

// Set from ParamSet
func (s Subspace) SetParamSet(ctx sdk.Context, ps ParamSet) {
	for _, pair := range ps.ParamSetPairs() {
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
