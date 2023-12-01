package ffi

// #include <stdint.h>
import "C"
import (
	"sync"
	"unsafe"

	"cosmossdk.io/core/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//export cosmossdk_store_open
func cosmossdk_store_open(moduleId uint32, ctxId uint32) uint32 {
	m := resolveModule(uint32(moduleId))
	ctx := resolveContext(uint32(ctxId))
	st := m.kvStoreService.OpenKVStore(ctx.ctx)
	storeId := nextStoreId()
	storeMap.Store(storeId, st)
	return uint32(codes.OK)
}

//export cosmossdk_store_has
func cosmossdk_store_has(storeId uint32, keyPtr *C.uint8_t, keyLen C.size_t) uint32 {
	st := resolveStore(storeId)
	key := unsafe.Slice(keyPtr, keyLen)
	has, err := st.Has(key)
	if err != nil {
		return uint32(status.Code(err))
	}
	if has {
		return uint32(codes.OK)
	} else {
		return uint32(codes.NotFound)
	}
}

//export cosmossdk_store_get
func cosmossdk_store_get(storeId C.uint32_t, keyPtr *C.uint8_t, keyLen C.size_t, valuePtr *C.uint8_t, valueLen *C.size_t) C.int32_t {
	st := resolveStore(storeId)
	key := unsafe.Slice(keyPtr, keyLen)
	value, err := st.Get(key)
	if err != nil {
		return status.Code(err)
	}
	if value != nil {
		n := len(value)
		allocated := *valueLen
		if n > allocated {
			return codes.ResourceExhausted
		}
		valuePtrSlice := unsafe.Slice(valuePtr, n)
		copy(valuePtrSlice, value)
		return codes.OK
	} else {
		return codes.NotFound
	}
}

//export cosmossdk_store_set
func cosmossdk_store_set(storeId C.uint32_t, keyPtr *C.uint8_t, keyLen C.size_t, valuePtr *C.uint8_t, valueLen C.size_t) C.int32_t {
	st := resolveStore(storeId)
	key := unsafe.Slice(keyPtr, keyLen)
	value := unsafe.Slice(valuePtr, valueLen)
	err := st.Set(key, value)
	return status.Code(err)
}

//export cosmossdk_store_delete
func cosmossdk_store_delete(storeId C.uint32_t, keyPtr *C.uint8_t, keyLen C.size_t) C.int32_t {
	st := resolveStore(storeId)
	key := unsafe.Slice(keyPtr, keyLen)
	err := st.Delete(key)
	return status.Code(err)
}

//export cosmossdk_store_close
func cosmossdk_store_close(storeId C.uint32_t) C.int32_t {
	storeMap.Delete(storeId)
	return codes.OK
}

func resolveStore(storeId uint32) store.KVStore {
	s, ok := storeMap.Load(storeId)
	if !ok {
		panic("invalid store")
	}
	return s.(store.KVStore)
}

func nextStoreId() interface{} {
	lastStoreId++
	return lastStoreId
}

var storeMap = &sync.Map{}
var lastStoreId uint32 = 0
