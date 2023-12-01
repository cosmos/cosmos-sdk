package wasmtime

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
//
// typedef int32_t (*allocations_t)();
// int32_t __allocations(void* f) {
//     return ((allocations_t)f)();
// }
import "C"
import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
)

type FFIModule struct {
	Handle      unsafe.Pointer
	exec        unsafe.Pointer
	execMsgSend unsafe.Pointer
	alloc       unsafe.Pointer
	free        unsafe.Pointer
}

func (f FFIModule) Alloc(n int) unsafe.Pointer {
	return unsafe.Pointer(C.__alloc(f.alloc, C.size_t(n)))
}

func (f FFIModule) Free(ptr unsafe.Pointer, n int) {
	C.__free(f.free, (*C.uint8_t)(ptr), C.size_t(n))
}

func (f FFIModule) Exec(input []byte) []byte {
	return f.doExec(f.exec, input)
}

func (f FFIModule) ExecMsgSend(input []byte) []byte {
	return f.doExec(f.execMsgSend, input)
}

func (f FFIModule) doExec(ptr unsafe.Pointer, input []byte) []byte {
	var outLen C.size_t
	out := C.exec(ptr, (*C.uint8_t)(&input[0]), C.size_t(len(input)), &outLen)
	res := C.GoBytes(unsafe.Pointer(out), C.int(outLen))
	f.Free(unsafe.Pointer(out), int(outLen))
	return res
}

func (f FFIModule) Allocations() int32 {
	allocationsPtr := C.dlsym(f.Handle, C.CString("allocations"))
	if allocationsPtr == nil {
		return -1
	}
	return int32(C.__allocations(allocationsPtr))
}

func LoadFFIModule(b testing.TB, path string) FFIModule {
	m := FFIModule{}
	m.Handle = C.dlopen(C.CString(path), C.RTLD_LAZY)
	require.NotNil(b, m.Handle)
	m.exec = C.dlsym(m.Handle, C.CString("exec"))
	require.NotNil(b, m.exec)
	m.execMsgSend = C.dlsym(m.Handle, C.CString("exec_msg_send"))
	require.NotNil(b, m.execMsgSend)
	m.alloc = C.dlsym(m.Handle, C.CString("__alloc"))
	require.NotNil(b, m.alloc)
	m.free = C.dlsym(m.Handle, C.CString("__free"))
	require.NotNil(b, m.free)
	return m
}
