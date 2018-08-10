package params

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

const (
	Gov    = "gov"
	Global = "global"
)

// Getter returns readonly struct
func (k Keeper) Getter() GetterProxy {
	return GetterProxy{Getter{k}}
}

// Setter returns read/write struct
func (k Keeper) Setter() SetterProxy {
	return SetterProxy{k.Getter(), Setter{Getter{k}}}
}

type GetterProxy struct {
	getter Getter
}

func getStoreKey(key string) string {
	if strings.HasPrefix(key, Gov+"/") || strings.HasPrefix(key, Global+"/") {
		return key
	}
	return fmt.Sprintf("%s/%s", Global, key)
}

func (proxy GetterProxy) Get(ctx sdk.Context, key string, res interface{}) (err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.Get(ctx, paramKey, res)
}

// GetRaw exposes getRaw
func (proxy GetterProxy) GetRaw(ctx sdk.Context, key string) []byte {
	paramKey := getStoreKey(key)
	return proxy.getter.GetRaw(ctx, paramKey)
}

// GetString is helper function for string params
func (proxy GetterProxy) GetString(ctx sdk.Context, key string) (res string, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetString(ctx, paramKey)
}

// GetBool is helper function for bool params
func (proxy GetterProxy) GetBool(ctx sdk.Context, key string) (res bool, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetBool(ctx, paramKey)
}

// GetInt16 is helper function for int16 params
func (proxy GetterProxy) GetInt16(ctx sdk.Context, key string) (res int16, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt16(ctx, paramKey)
}

// GetInt32 is helper function for int32 params
func (proxy GetterProxy) GetInt32(ctx sdk.Context, key string) (res int32, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt32(ctx, paramKey)
}

// GetInt64 is helper function for int64 params
func (proxy GetterProxy) GetInt64(ctx sdk.Context, key string) (res int64, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt64(ctx, paramKey)
}

// GetUint16 is helper function for uint16 params
func (proxy GetterProxy) GetUint16(ctx sdk.Context, key string) (res uint16, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint16(ctx, paramKey)
}

// GetUint32 is helper function for uint32 params
func (proxy GetterProxy) GetUint32(ctx sdk.Context, key string) (res uint32, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint32(ctx, paramKey)
}

// GetUint64 is helper function for uint64 params
func (proxy GetterProxy) GetUint64(ctx sdk.Context, key string) (res uint64, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint64(ctx, paramKey)
}

// GetInt is helper function for sdk.Int params
func (proxy GetterProxy) GetInt(ctx sdk.Context, key string) (res sdk.Int, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt(ctx, paramKey)
}

// GetUint is helper function for sdk.Uint params
func (proxy GetterProxy) GetUint(ctx sdk.Context, key string) (res sdk.Uint, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint(ctx, paramKey)
}

// GetRat is helper function for rat params
func (proxy GetterProxy) GetRat(ctx sdk.Context, key string) (res sdk.Rat, err error) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetRat(ctx, paramKey)
}

// GetStringWithDefault is helper function for string params with default value
func (proxy GetterProxy) GetStringWithDefault(ctx sdk.Context, key string, def string) (res string) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetStringWithDefault(ctx, paramKey, def)
}

// GetBoolWithDefault is helper function for bool params with default value
func (proxy GetterProxy) GetBoolWithDefault(ctx sdk.Context, key string, def bool) (res bool) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetBoolWithDefault(ctx, paramKey, def)
}

// GetInt16WithDefault is helper function for int16 params with default value
func (proxy GetterProxy) GetInt16WithDefault(ctx sdk.Context, key string, def int16) (res int16) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt16WithDefault(ctx, paramKey, def)
}

// GetInt32WithDefault is helper function for int32 params with default value
func (proxy GetterProxy) GetInt32WithDefault(ctx sdk.Context, key string, def int32) (res int32) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt32WithDefault(ctx, paramKey, def)
}

// GetInt64WithDefault is helper function for int64 params with default value
func (proxy GetterProxy) GetInt64WithDefault(ctx sdk.Context, key string, def int64) (res int64) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetInt64WithDefault(ctx, paramKey, def)
}

// GetUint16WithDefault is helper function for uint16 params with default value
func (proxy GetterProxy) GetUint16WithDefault(ctx sdk.Context, key string, def uint16) (res uint16) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint16WithDefault(ctx, paramKey, def)
}

// GetUint32WithDefault is helper function for uint32 params with default value
func (proxy GetterProxy) GetUint32WithDefault(ctx sdk.Context, key string, def uint32) (res uint32) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint32WithDefault(ctx, paramKey, def)
}

// GetUint64WithDefault is helper function for uint64 params with default value
func (proxy GetterProxy) GetUint64WithDefault(ctx sdk.Context, key string, def uint64) (res uint64) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUint64WithDefault(ctx, paramKey, def)
}

// GetIntWithDefault is helper function for sdk.Int params with default value
func (proxy GetterProxy) GetIntWithDefault(ctx sdk.Context, key string, def sdk.Int) (res sdk.Int) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetIntWithDefault(ctx, paramKey, def)
}

// GetUintWithDefault is helper function for sdk.Uint params with default value
func (proxy GetterProxy) GetUintWithDefault(ctx sdk.Context, key string, def sdk.Uint) (res sdk.Uint) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetUintWithDefault(ctx, paramKey, def)
}

// GetRatWithDefault is helper function for sdk.Rat params with default value
func (proxy GetterProxy) GetRatWithDefault(ctx sdk.Context, key string, def sdk.Rat) (res sdk.Rat) {
	paramKey := getStoreKey(key)
	return proxy.getter.GetRatWithDefault(ctx, paramKey, def)
}

type SetterProxy struct {
	GetterProxy
	Setter
}

// Set exposes set
func (proxy SetterProxy) Set(ctx sdk.Context, key string, param interface{}) error {
	paramKey := getStoreKey(key)
	return proxy.Setter.Set(ctx, paramKey, param)
}

// SetRaw exposes setRaw
func (proxy SetterProxy) SetRaw(ctx sdk.Context, key string, param []byte) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetRaw(ctx, paramKey, param)
}

// SetString is helper function for string params
func (proxy SetterProxy) SetString(ctx sdk.Context, key string, param string) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetString(ctx, paramKey, param)
}

// SetBool is helper function for bool params
func (proxy SetterProxy) SetBool(ctx sdk.Context, key string, param bool) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetBool(ctx, paramKey, param)
}

// SetInt16 is helper function for int16 params
func (proxy SetterProxy) SetInt16(ctx sdk.Context, key string, param int16) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetInt16(ctx, paramKey, param)
}

// SetInt32 is helper function for int32 params
func (proxy SetterProxy) SetInt32(ctx sdk.Context, key string, param int32) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetInt32(ctx, paramKey, param)
}

// SetInt64 is helper function for int64 params
func (proxy SetterProxy) SetInt64(ctx sdk.Context, key string, param int64) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetInt64(ctx, paramKey, param)
}

// SetUint16 is helper function for uint16 params
func (proxy SetterProxy) SetUint16(ctx sdk.Context, key string, param uint16) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetUint16(ctx, paramKey, param)
}

// SetUint32 is helper function for uint32 params
func (proxy SetterProxy) SetUint32(ctx sdk.Context, key string, param uint32) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetUint32(ctx, paramKey, param)
}

// SetUint64 is helper function for uint64 params
func (proxy SetterProxy) SetUint64(ctx sdk.Context, key string, param uint64) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetUint64(ctx, paramKey, param)
}

// SetInt is helper function for sdk.Int params
func (proxy SetterProxy) SetInt(ctx sdk.Context, key string, param sdk.Int) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetInt(ctx, paramKey, param)
}

// SetUint is helper function for sdk.Uint params
func (proxy SetterProxy) SetUint(ctx sdk.Context, key string, param sdk.Uint) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetUint(ctx, paramKey, param)
}

// SetRat is helper function for rat params
func (proxy SetterProxy) SetRat(ctx sdk.Context, key string, param sdk.Rat) {
	paramKey := getStoreKey(key)
	proxy.Setter.SetRat(ctx, paramKey, param)
}
