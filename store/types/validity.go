package types

import "errors"

var (
	// 128K - 1
	MaxKeyLength = (1 << 17) - 1
	// 2G - 1
	MaxValueLength = (1 << 31) - 1
)

// AssertValidKey checks if the key is valid(key is not nil, not empty and within length limit)
func AssertValidKey(key []byte) {
	if len(key) == 0 {
		panic("key is nil or empty")
	}
	if len(key) > MaxKeyLength {
		panic("key is too large")
	}
}

// AssertValidValue checks if the value is valid(value is not nil and within length limit)
func AssertValidValue(value []byte) {
	if value == nil {
		panic("value is nil")
	}
	AssertValidValueLength(len(value))
}

// AssertValidValueGeneric checks if the value is valid(value is not nil and within length limit)
func AssertValidValueGeneric[V any](value V, isZero func(V) bool, valueLen func(V) int) {
	if isZero(value) {
		panic("value is nil")
	}
	AssertValidValueLength(valueLen(value))
}

// AssertValidValueLength checks if the value length is within length limit
func AssertValidValueLength(l int) {
	if l > MaxValueLength {
		panic(errors.New("value is too large"))
	}
}
