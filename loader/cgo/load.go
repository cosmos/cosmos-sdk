package cgo

// #cgo LDFLAGS: -ldl
// #include <dlfcn.h>
// #include <stdint.h>
//
// typedef const uint8_t* (*proto_file_descriptor_set_t)(size_t* out_len);
// const uint8_t* proto_file_descriptor_set(void* f, size_t* out_len) {
//     return ((proto_file_descriptor_set_t)f)(out_len);
// }
//
//typedef struct ModuleInitializer {
//  const char *name;
//  const uint8_t **proto_files;
//} ModuleInitializer;
//
//typedef struct ModuleInitializers {
//	size_t count;
//	const struct ModuleInitializer *initializers;
//} ModuleInitializers;
//
//typedef const ModuleInitializers* (*load_modules_t)();
//const ModuleInitializers* load_modules(void* f) {
//    return ((load_modules_t)f)();
//}
//
// typedef const uint8_t* (*unary_method_t)(const uintptr_t, const uint8_t* input, size_t len, size_t* out_len);
// const uint8_t* unary_method(void* f, const uintptr_t ctx, const uint8_t* input, size_t len, size_t* out_len) {
//     return ((unary_method_t)f)(ctx, input, len, out_len);
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
	"fmt"
	"unsafe"
)

func LoadLibrary(path string) {
	err := loadLibrary(path)
	if err != nil {
		panic(err)
	}
}

func loadLibrary(path string) error {
	lib := C.dlopen(C.CString(path), C.RTLD_LAZY)
	if lib == nil {
		return fmt.Errorf("failed to load library %s", path)
	}
	protoFileDescriptorSet := C.dlsym(lib, C.CString("__proto_file_descriptor_set"))
	if protoFileDescriptorSet == nil {
		return fmt.Errorf("failed to load proto_file_descriptor_set from %s", path)
	}
	var outLen C.size_t
	protoFileDescriptorSetBytes := C.proto_file_descriptor_set(protoFileDescriptorSet, &outLen)
	fmt.Printf("protoFileDescriptorSetBytes: %v\n", C.GoBytes(unsafe.Pointer(protoFileDescriptorSetBytes), C.int(outLen)))

	entry := C.dlsym(lib, C.CString("__entry"))
	if entry == nil {
		return fmt.Errorf("failed to load entry from %s", path)
	}

	//loadModules := C.dlsym(lib, C.CString("__load_modules"))
	//mods := C.load_modules(loadModules)
	//for i := 0; i < int(mods.count); i++ {
	//	initializer := mods.initializers[i]
	//	name := C.GoString(initializer.name)
	//	fmt.Printf("Loading module %s\n", name)
	//	//appmodule.Register(nil,
	//	//	appmodule.Provide(func(kvStoreService store.KVStoreService) appmodule.AppModule {
	//	//		return module{
	//	//			kvStoreService: kvStoreService,
	//	//		}
	//	//	}))
	//}
	// for each module

	// load library
	// load FileDescriptorSet
	// load module names & initializers

	// load module
	// call initializer with: config, store, client
	return nil
}
