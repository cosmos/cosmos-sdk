package params

import (
	"fmt"
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// Keeper manages global parameter store
type Keeper struct {
	cdc *wire.Codec
	key sdk.StoreKey
}

// NewKeeper constructs a new Keeper
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey) Keeper {
	return Keeper{
		cdc: cdc,
		key: key,
	}
}

// InitKeeper constructs a new Keeper with initial parameters
func InitKeeper(ctx sdk.Context, cdc *wire.Codec, key sdk.StoreKey, params ...interface{}) Keeper {
	if len(params)%2 != 0 {
		panic("Odd params list length for InitKeeper")
	}

	k := NewKeeper(cdc, key)

	for i := 0; i < len(params); i += 2 {
		k.set(ctx, params[i].(string), params[i+1])
	}

	return k
}

// get automatically unmarshalls parameter to pointer
func (k Keeper) get(ctx sdk.Context, key string, ptr interface{}) error {
	store := ctx.KVStore(k.key)
	bz := store.Get([]byte(key))
	return k.cdc.UnmarshalBinary(bz, ptr)
}

// getRaw returns raw byte slice
func (k Keeper) getRaw(ctx sdk.Context, key string) []byte {
	store := ctx.KVStore(k.key)
	return store.Get([]byte(key))
}

// set automatically marshalls and type check parameter
func (k Keeper) set(ctx sdk.Context, key string, param interface{}) error {
	store := ctx.KVStore(k.key)
	bz := store.Get([]byte(key))
	if bz != nil {
		ptrty := reflect.PtrTo(reflect.TypeOf(param))
		ptr := reflect.New(ptrty).Interface()

		if k.cdc.UnmarshalBinary(bz, ptr) != nil {
			return fmt.Errorf("Type mismatch with stored param and provided param")
		}
	}

	bz, err := k.cdc.MarshalBinary(param)
	if err != nil {
		return err
	}
	store.Set([]byte(key), bz)

	return nil
}

// setRaw sets raw byte slice
func (k Keeper) setRaw(ctx sdk.Context, key string, param []byte) {
	store := ctx.KVStore(k.key)
	store.Set([]byte(key), param)
}

// Getter returns readonly struct
func (k Keeper) Getter() Getter {
	return Getter{k}
}

// Setter returns read/write struct
func (k Keeper) Setter() Setter {
	return Setter{Getter{k}}
}

// Getter exposes methods related with only getting params
type Getter struct {
	k Keeper
}

// Get exposes get
func (k Getter) Get(ctx sdk.Context, key string, ptr interface{}) error {
	return k.k.get(ctx, key, ptr)
}

// GetRaw exposes getRaw
func (k Getter) GetRaw(ctx sdk.Context, key string) []byte {
	return k.k.getRaw(ctx, key)
}

// GetString is helper function for string params
func (k Getter) GetString(ctx sdk.Context, key string) (res string, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetBool is helper function for bool params
func (k Getter) GetBool(ctx sdk.Context, key string) (res bool, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetInt16 is helper function for int16 params
func (k Getter) GetInt16(ctx sdk.Context, key string) (res int16, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetInt32 is helper function for int32 params
func (k Getter) GetInt32(ctx sdk.Context, key string) (res int32, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetInt64 is helper function for int64 params
func (k Getter) GetInt64(ctx sdk.Context, key string) (res int64, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetUint16 is helper function for uint16 params
func (k Getter) GetUint16(ctx sdk.Context, key string) (res uint16, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetUint32 is helper function for uint32 params
func (k Getter) GetUint32(ctx sdk.Context, key string) (res uint32, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetUint64 is helper function for uint64 params
func (k Getter) GetUint64(ctx sdk.Context, key string) (res uint64, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetInt is helper function for sdk.Int params
func (k Getter) GetInt(ctx sdk.Context, key string) (res sdk.Int, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetUint is helper function for sdk.Uint params
func (k Getter) GetUint(ctx sdk.Context, key string) (res sdk.Uint, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetRat is helper function for rat params
func (k Getter) GetRat(ctx sdk.Context, key string) (res sdk.Rat, err error) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	err = k.k.cdc.UnmarshalBinary(bz, &res)
	return
}

// GetStringWithDefault is helper function for string params with default value
func (k Getter) GetStringWithDefault(ctx sdk.Context, key string, def string) (res string) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetBoolWithDefault is helper function for bool params with default value
func (k Getter) GetBoolWithDefault(ctx sdk.Context, key string, def bool) (res bool) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetInt16WithDefault is helper function for int16 params with default value
func (k Getter) GetInt16WithDefault(ctx sdk.Context, key string, def int16) (res int16) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetInt32WithDefault is helper function for int32 params with default value
func (k Getter) GetInt32WithDefault(ctx sdk.Context, key string, def int32) (res int32) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetInt64WithDefault is helper function for int64 params with default value
func (k Getter) GetInt64WithDefault(ctx sdk.Context, key string, def int64) (res int64) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetUint16WithDefault is helper function for uint16 params with default value
func (k Getter) GetUint16WithDefault(ctx sdk.Context, key string, def uint16) (res uint16) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetUint32WithDefault is helper function for uint32 params with default value
func (k Getter) GetUint32WithDefault(ctx sdk.Context, key string, def uint32) (res uint32) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetUint64WithDefault is helper function for uint64 params with default value
func (k Getter) GetUint64WithDefault(ctx sdk.Context, key string, def uint64) (res uint64) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetIntWithDefault is helper function for sdk.Int params with default value
func (k Getter) GetIntWithDefault(ctx sdk.Context, key string, def sdk.Int) (res sdk.Int) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetUintWithDefault is helper function for sdk.Uint params with default value
func (k Getter) GetUintWithDefault(ctx sdk.Context, key string, def sdk.Uint) (res sdk.Uint) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// GetRatWithDefault is helper function for sdk.Rat params with default value
func (k Getter) GetRatWithDefault(ctx sdk.Context, key string, def sdk.Rat) (res sdk.Rat) {
	store := ctx.KVStore(k.k.key)
	bz := store.Get([]byte(key))
	if bz == nil {
		return def
	}
	k.k.cdc.MustUnmarshalBinary(bz, &res)
	return
}

// Setter exposes all methods including Set
type Setter struct {
	Getter
}

// Set exposes set
func (k Setter) Set(ctx sdk.Context, key string, param interface{}) error {
	return k.k.set(ctx, key, param)
}

// SetRaw exposes setRaw
func (k Setter) SetRaw(ctx sdk.Context, key string, param []byte) {
	k.k.setRaw(ctx, key, param)
}

// SetString is helper function for string params
func (k Setter) SetString(ctx sdk.Context, key string, param string) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetBool is helper function for bool params
func (k Setter) SetBool(ctx sdk.Context, key string, param bool) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetInt16 is helper function for int16 params
func (k Setter) SetInt16(ctx sdk.Context, key string, param int16) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetInt32 is helper function for int32 params
func (k Setter) SetInt32(ctx sdk.Context, key string, param int32) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetInt64 is helper function for int64 params
func (k Setter) SetInt64(ctx sdk.Context, key string, param int64) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetUint16 is helper function for uint16 params
func (k Setter) SetUint16(ctx sdk.Context, key string, param uint16) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetUint32 is helper function for uint32 params
func (k Setter) SetUint32(ctx sdk.Context, key string, param uint32) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetUint64 is helper function for uint64 params
func (k Setter) SetUint64(ctx sdk.Context, key string, param uint64) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetInt is helper function for sdk.Int params
func (k Setter) SetInt(ctx sdk.Context, key string, param sdk.Int) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetUint is helper function for sdk.Uint params
func (k Setter) SetUint(ctx sdk.Context, key string, param sdk.Uint) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}

// SetRat is helper function for rat params
func (k Setter) SetRat(ctx sdk.Context, key string, param sdk.Rat) {
	if err := k.k.set(ctx, key, param); err != nil {
		panic(err)
	}
}
