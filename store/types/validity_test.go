package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/store/types"
)

func TestAssertValidKey(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() { types.AssertValidKey([]byte{0x01}) })
	require.Panics(t, func() { types.AssertValidKey([]byte{}) })
	require.Panics(t, func() { types.AssertValidKey(nil) })
}

func TestAssertValidValue(t *testing.T) {
	t.Parallel()
	require.NotPanics(t, func() { types.AssertValidValue([]byte{}) })
	require.NotPanics(t, func() { types.AssertValidValue([]byte{0x01}) })
	require.Panics(t, func() { types.AssertValidValue(nil) })
}

func TestAssertValidValueGeneric(t *testing.T) {
	t.Parallel()
	bytesIsZero := func(b []byte) bool { return b == nil }
	bytesValueLen := func(b []byte) int { return len(b) }
	require.NotPanics(t, func() {
		types.AssertValidValueGeneric(
			[]byte{},
			bytesIsZero,
			bytesValueLen,
		)
	})
	require.NotPanics(t, func() {
		types.AssertValidValueGeneric(
			[]byte{0x1},
			bytesIsZero,
			bytesValueLen,
		)
	})
	require.Panics(t, func() {
		types.AssertValidValueGeneric(
			nil,
			bytesIsZero,
			bytesValueLen,
		)
	})
	require.Panics(t, func() {
		types.AssertValidValueGeneric(
			make([]byte, types.MaxValueLength+1),
			bytesIsZero,
			bytesValueLen,
		)
	})
}
