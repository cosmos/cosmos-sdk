package cgo

// #cgo LDFLAGS: -ldl
// #include <dlfcn.h>
// #include <stdint.h>
// #include "../../rust/cosmossdk_core/cosmossdk_core.h"
//
// typedef const InitData* (*init_t)();
// const InitData* init(void* f) {
// 	return ((init_t)f)();
// }
import "C"
import (
	"fmt"
)

func LoadLibrary(path string) {
	err := loadLibrary(path)
	if err != nil {
		panic(err)
	}
}

func loadLibrary(path string) error {
	lib := C.dlopen(C.CString(path), C.RTLD_LAZY|C.RTLD_LOCAL)
	if lib == nil {
		return fmt.Errorf("failed to load library %s", path)
	}

	init := C.dlsym(lib, C.CString("__init"))
	if init == nil {
		return fmt.Errorf("failed to load __init from %s", path)
	}

	initData := C.init(init)
	fmt.Printf("initData: %v\n", initData)

	//protoFileDescriptorSet := C.dlsym(lib, C.CString("__proto_file_descriptor_set"))
	//if protoFileDescriptorSet == nil {
	//	return fmt.Errorf("failed to load proto_file_descriptor_set from %s", path)
	//}
	//var outLen C.size_t
	//protoFileDescriptorSetBytes := C.proto_file_descriptor_set(protoFileDescriptorSet, &outLen)
	//fmt.Printf("protoFileDescriptorSetBytes: %v\n", C.GoBytes(unsafe.Pointer(protoFileDescriptorSetBytes), C.int(outLen)))
	//
	//entry := C.dlsym(lib, C.CString("__entry"))
	//if entry == nil {
	//	return fmt.Errorf("failed to load entry from %s", path)
	//}

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
