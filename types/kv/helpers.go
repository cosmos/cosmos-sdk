package kv

import "fmt"

func AssertKeyAtLeastLength(bz []byte, length int) {
	if len(bz) < length {
		panic(fmt.Sprintf("expected key of length at least %d, got %d", length, len(bz)))
	}
}

func AssertKeyLength(bz []byte, length int) {
	if len(bz) != length {
		panic(fmt.Sprintf("unexpected key length; got: %d, expected: %d", len(bz), length))
	}
}
