package cgo

import "C"
import (
	"fmt"
	"unsafe"

	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

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
	"errors"

	"google.golang.org/protobuf/proto"
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

	// load proto file descriptors
	fdsBz := C.GoBytes(unsafe.Pointer(initData.proto_file_descriptors), C.int(initData.proto_file_descriptors_len))
	err := loadFileDescriptors(fdsBz, protoregistry.GlobalFiles)
	if err != nil {
		return err
	}

	// register modules
	modDescriptors := unsafe.Slice(initData.module_descriptors, int(initData.num_modules))
	for _, mod := range modDescriptors {
		name := string(C.GoBytes(unsafe.Pointer(mod.name), C.int(mod.name_len)))
		fmt.Printf("loading module %s\n", name)
	}
	//for i := 0; i < int(numModules); i++ {
	//	mod := initData.module_descriptors[i]
	//	fmt.Printf("loading module %+v\n", mod)
	//	//fmt.Printf("loading module %s\n", C.GoString(name))
	//}

	return nil
}

func loadFileDescriptors(fdsBz []byte, registry *protoregistry.Files) error {
	fds := &descriptorpb.FileDescriptorSet{}
	err := proto.Unmarshal(fdsBz, fds)
	if err != nil {
		return err
	}

	for _, fd := range fds.File {
		existing, err := registry.FindFileByPath(fd.GetName())
		if err != nil && !errors.Is(err, protoregistry.NotFound) {
			return err
		}
		if existing != nil {
			// TODO: compare existing and fd
			continue
		}
		file, err := protodesc.NewFile(fd, registry)
		if err != nil {
			return err
		}

		err = registry.RegisterFile(file)
		if err != nil {
			return err
		}
	}
	return nil
}
