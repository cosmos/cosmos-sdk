package cgo

import "C"

//export cosmossdk_invoke
func cosmossdk_invoke(contextPtr uintptr, methodPtr uintptr, reqPtr uintptr, reqLen uintptr, resPtr uintptr, resLen uintptr) uint32 {
	//ctx := cgo.Handle(contextPtr).Value().(*contextWrapper)
	panic("implement me")
}
