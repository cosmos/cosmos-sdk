package wasm

import "C"
import (
	"github.com/stretchr/testify/require"
	"unsafe"
)

// #cgo LDFLAGS: -ldl
// #include <dlfcn.h>
// #include <stdint.h>
//
// typedef const uint8_t* (*exec_t)(const uint8_t* input, size_t len, size_t* out_len);
// const uint8_t* exec(void* f, const uint8_t* input, size_t len, size_t* out_len) {
//     return ((exec_t)f)(input, len, out_len);
// }
//
// typedef uint8_t* (*alloc_t)(size_t size);
// uint8_t* __alloc(void* f, size_t size) {
//     return ((alloc_t)f)(size);
// }
//
// typedef void (*free_t)(uint8_t* ptr, size_t size);
// void __free(void* f, uint8_t* ptr, size_t size) {
//     ((free_t)f)(ptr, size);
// }
import "C"
import (
	"testing"
)

type FFIModule struct {
	Handle   unsafe.Pointer
	ExecPtr  unsafe.Pointer
	AllocPtr unsafe.Pointer
	FreePtr  unsafe.Pointer
}

func (f FFIModule) Alloc(n int) unsafe.Pointer {
	return unsafe.Pointer(C.__alloc(f.AllocPtr, C.size_t(n)))
}

func (f FFIModule) Free(ptr unsafe.Pointer, n int) {
	C.__free(f.FreePtr, (*C.uint8_t)(ptr), C.size_t(n))
}

func (f FFIModule) Exec(input []byte) []byte {
	var outLen C.size_t
	out := C.exec(f.ExecPtr, (*C.uint8_t)(&input[0]), C.size_t(len(input)), &outLen)
	return C.GoBytes(unsafe.Pointer(out), C.int(outLen))
}

func LoadFFIModule(b testing.TB, path string) FFIModule {
	m := FFIModule{}
	m.Handle = C.dlopen(C.CString(path), C.RTLD_LAZY)
	require.NotNil(b, m.Handle)
	m.ExecPtr = C.dlsym(m.Handle, C.CString("exec"))
	require.NotNil(b, m.ExecPtr)
	m.AllocPtr = C.dlsym(m.Handle, C.CString("__alloc"))
	require.NotNil(b, m.AllocPtr)
	m.FreePtr = C.dlsym(m.Handle, C.CString("__free"))
	require.NotNil(b, m.FreePtr)
	return m
}
