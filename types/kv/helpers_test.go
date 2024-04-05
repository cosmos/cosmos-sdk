package kv_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/stretchr/testify/assert"
)

func TestAssertKeyAtLeastLength(t *testing.T) {
	cases := []struct {
		name        string
		key         []byte
		length      int
		expectPanic bool
	}{
		{
			name:        "Store key is less then the given length",
			key:         []byte("hello"),
			length:      10,
			expectPanic: true,
		},
		{
			name:        "Store key is equal to the given length",
			key:         []byte("store-key"),
			length:      9,
			expectPanic: false,
		},
		{
			name:        "Store key is greater then the given length",
			key:         []byte("unique"),
			length:      3,
			expectPanic: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tc.expectPanic {
					assert.NotNil(t, r)
					return
				}
				assert.Nil(t, r)
			}()
			kv.AssertKeyAtLeastLength(tc.key, tc.length)
		})
	}
}

func TestAssertKeyLength(t *testing.T) {
	cases := []struct {
		name        string
		key         []byte
		length      int
		expectPanic bool
	}{
		{
			name:        "Store key is less then the given length",
			key:         []byte("hello"),
			length:      10,
			expectPanic: true,
		},
		{
			name:        "Store key is equal to the given length",
			key:         []byte("store-key"),
			length:      9,
			expectPanic: false,
		},
		{
			name:        "Store key is greater then the given length",
			key:         []byte("unique"),
			length:      3,
			expectPanic: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tc.expectPanic {
					assert.NotNil(t, r)
					return
				}
				assert.Nil(t, r)
			}()
			kv.AssertKeyLength(tc.key, tc.length)
		})
	}
}
