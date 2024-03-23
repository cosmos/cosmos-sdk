package cgo

import "testing"

func Test1(t *testing.T) {
	LoadLibrary("../../rust/target/release/libcounter_pb.dylib")
}
