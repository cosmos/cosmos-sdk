package kv

import "fmt"

func AssertKeyAtLeastLength(bz []byte, length int) {
	if len(bz) < length {
		panic(fmt.Sprintf("expected key of length at least %d, got %d", length, len(bz)))
	}
}
