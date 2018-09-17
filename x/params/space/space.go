package space

import (
	"fmt"
	"reflect"

	tmlibs "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Individual parameter store for each keeper
type Space struct {
	cdc  *codec.Codec
	key  sdk.StoreKey
	tkey sdk.StoreKey

	space []byte
}

// NewSpace constructs a store with namespace
func NewSpace(cdc *codec.Codec, key sdk.StoreKey, tkey sdk.StoreKey, space string) Space {
	if !tmlibs.IsASCIIText(space) {
		panic("paramstore space expressions can only contain alphanumeric characters")
	}

	return Space{
		cdc:  cdc,
		key:  key,
		tkey: tkey,

		space: append([]byte(space), '/'),
	}
}

// Get parameter from store
func (s Space) Get(ctx sdk.Context, key string, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.space)
	bz := store.Get([]byte(key))
	err := s.cdc.UnmarshalJSON(bz, ptr)
	if err != nil {
		panic(err)
	}
}

// GetIfExists do not modify ptr if the stored parameter is nil
func (s Space) GetIfExists(ctx sdk.Context, key string, ptr interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.space)
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
func (s Space) GetRaw(ctx sdk.Context, key string) []byte {
	store := ctx.KVStore(s.key).Prefix(s.space)
	res := store.Get([]byte(key))
	return res
}

// Check if the parameter is set in the store
func (s Space) Has(ctx sdk.Context, key string) bool {
	store := ctx.KVStore(s.key).Prefix(s.space)
	return store.Has([]byte(key))
}

// Returns true if the parameter is set in the block
func (s Space) Modified(ctx sdk.Context, key string) bool {
	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	return tstore.Has([]byte(key))
}

// Set parameter, return error if stored parameter has different type from input
func (s Space) Set(ctx sdk.Context, key string, param interface{}) {
	store := ctx.KVStore(s.key).Prefix(s.space)
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

	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	tstore.Set(keybz, []byte{})
}

// Set raw bytes of parameter
func (s Space) SetRaw(ctx sdk.Context, key string, param []byte) {
	keybz := []byte(key)

	store := ctx.KVStore(s.key).Prefix(s.space)
	store.Set(keybz, param)

	tstore := ctx.KVStore(s.tkey).Prefix(s.space)
	tstore.Set(keybz, []byte{})
}

// Set from ParamStruct
func (s Space) SetFromParamStruct(ctx sdk.Context, ps ParamStruct) {
	for _, pair := range ps.KeyFieldPairs() {
		s.Set(ctx, pair.Key, pair.Field)
	}
}

// Returns a KVStore identical with the paramspace
func (s Space) KVStore(ctx sdk.Context) sdk.KVStore {
	return ctx.KVStore(s.key).Prefix(s.space)
}

// Wrapper of Space, provides immutable functions only
type ReadOnlySpace struct {
	s Space
}

// Exposes Get
func (ros ReadOnlySpace) Get(ctx sdk.Context, key string, ptr interface{}) {
	ros.s.Get(ctx, key, ptr)
}

// Exposes GetRaw
func (ros ReadOnlySpace) GetRaw(ctx sdk.Context, key string) []byte {
	return ros.s.GetRaw(ctx, key)
}

// Exposes Has
func (ros ReadOnlySpace) Has(ctx sdk.Context, key string) bool {
	return ros.s.Has(ctx, key)
}

// Exposes Modified
func (ros ReadOnlySpace) Modified(ctx sdk.Context, key string) bool {
	return ros.s.Modified(ctx, key)
}
