package types

import math "math"

const (
	MaxKeyLength   = math.MaxUint16
	MaxValueLength = math.MaxInt32
)

// AssertValidKey checks if the key is valid(key is not nil)
func AssertValidKey(key []byte) {
	if len(key) == 0 {
		panic("key is nil")
	}
	if len(key) > MaxKeyLength {
		panic("key is too large")
	}
}

// AssertValidValue checks if the value is valid(value is not nil)
func AssertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
	if len(value) > MaxValueLength {
		panic("value is too large")
	}
}
