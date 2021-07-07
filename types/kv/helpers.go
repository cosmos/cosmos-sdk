package kv

import "fmt"

func AssertKeyAtLeastLength(bz []byte, length int) {
	if len(bz) < length {
		panic(fmt.Sprintf("expected key of length at least %d, got %d", length, len(bz)))
	}
}

func AssertKeyLength(addr []byte, length int) {
	if len(addr) != length {
		panic(fmt.Sprintf("unexpected key length; got: %d, expected: %d", len(addr), length))
	}
}
