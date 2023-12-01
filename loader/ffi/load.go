package ffi

// #cgo LDFLAGS: -ldl
// #include <dlfcn.h>
// #include <stdint.h>
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

func LoadLibrary(path string) {
	//lib := C.dlopen(C.CString(path), C.RTLD_LAZY)
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
}
