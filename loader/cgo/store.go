package cgo

import "runtime/cgo"

// #include <stdint.h>
//
// typedef uint32_t store_has_t(uintptr_t, const uint8_t*, size_t);
// typedef uint32_t store_get_t(uintptr_t, const uint8_t*, size_t, uint8_t*, size_t*);
// typedef uint32_t store_set_t(uintptr_t, const uint8_t*, size_t, const uint8_t*, size_t);
// typedef uint32_t store_delete_t(uintptr_t, const uint8_t*, size_t);
// typedef uint32_t store_close_t(uintptr_t);
//
// typedef struct KVStoreService {
// 	store_has_t* has;
// 	store_get_t* get;
// 	store_set_t* set;
// 	store_delete_t* del;
// 	store_close_t* close;
// } KVStoreService;
import "C"
import (
	"unsafe"

	"cosmossdk.io/core/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//export cosmossdk_store_open
func cosmossdk_store_open(moduleId uintptr, ctxId uintptr) uintptr {
	m := cgo.Handle(moduleId).Value().(*module)
	ctx := cgo.Handle(ctxId).Value().(*contextWrapper)
	st := m.kvStoreService.OpenKVStore(ctx.ctx)
	return uintptr(cgo.NewHandle(st))
}

//export cosmossdk_store_has
func cosmossdk_store_has(storeId uintptr, keyPtr unsafe.Pointer, keyLen C.size_t) uint32 {
	key := C.GoBytes(keyPtr, C.int(keyLen))
	st := cgo.Handle(storeId).Value().(store.KVStore)
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
func cosmossdk_store_get(storeId C.uint32_t, keyPtr unsafe.Pointer, keyLen C.size_t, valuePtr unsafe.Pointer, valueLen *C.size_t) uint32 {
	st := cgo.Handle(storeId).Value().(store.KVStore)
	key := C.GoBytes(keyPtr, C.int(keyLen))
	value, err := st.Get(key)
	if err != nil {
		return uint32(status.Code(err))
	}
	if value != nil {
		n := len(value)
		allocated := int(*valueLen)
		if n > allocated {
			return uint32(codes.ResourceExhausted)
		}
		valuePtrSlice := C.GoBytes(valuePtr, C.int(n))
		copy(valuePtrSlice, value)
		return uint32(codes.OK)
	} else {
		return uint32(codes.NotFound)
	}
}

//export cosmossdk_store_set
func cosmossdk_store_set(storeId C.uint32_t, keyPtr unsafe.Pointer, keyLen C.size_t, valuePtr unsafe.Pointer, valueLen C.size_t) uint32 {
	st := cgo.Handle(storeId).Value().(store.KVStore)
	key := C.GoBytes(keyPtr, C.int(keyLen))
	value := C.GoBytes(valuePtr, C.int(valueLen))
	err := st.Set(key, value)
	return uint32(status.Code(err))
}

//export cosmossdk_store_delete
func cosmossdk_store_delete(storeId C.uint32_t, keyPtr unsafe.Pointer, keyLen C.size_t) uint32 {
	st := cgo.Handle(storeId).Value().(store.KVStore)
	key := C.GoBytes(keyPtr, C.int(keyLen))
	err := st.Delete(key)
	return uint32(status.Code(err))
}

//export cosmossdk_store_close
func cosmossdk_store_close(storeId C.uint32_t) uint32 {
	cgo.Handle(storeId).Delete()
	return uint32(codes.OK)
}
