package internal

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnsafeBytes(t *testing.T) {
	hello := []byte("hello")
	unsafe := WrapUnsafeBytes(hello)
	require.False(t, unsafe.IsNil())
	require.Equal(t, hello, unsafe.UnsafeBytes())
	safeCopy := unsafe.SafeCopy()
	require.Equal(t, hello, safeCopy)
	require.NotSame(t, &hello[0], &safeCopy[0]) // different underlying array

	safe := WrapSafeBytes(hello)
	require.False(t, safe.IsNil())
	require.Equal(t, hello, safe.UnsafeBytes())
	safeCopy2 := safe.SafeCopy()
	require.Equal(t, hello, safeCopy2)
	require.Same(t, &hello[0], &safeCopy2[0]) // same underlying array

	nilUnsafe := WrapUnsafeBytes(nil)
	require.True(t, nilUnsafe.IsNil())
	require.Nil(t, nilUnsafe.UnsafeBytes())
	require.Nil(t, nilUnsafe.SafeCopy())

	nilSafe := WrapSafeBytes(nil)
	require.True(t, nilSafe.IsNil())
	require.Nil(t, nilSafe.UnsafeBytes())
	require.Nil(t, nilSafe.SafeCopy())

	nilInit := UnsafeBytes{}
	require.True(t, nilInit.IsNil())
	require.Nil(t, nilInit.UnsafeBytes())
	require.Nil(t, nilInit.SafeCopy())
}
