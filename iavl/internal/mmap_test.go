package internal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMmap_ReadWrite(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "test")
	require.NoError(t, err)
	defer f.Close()
	_, err = f.Write([]byte("hello world"))
	require.NoError(t, err)

	m, err := NewMmap(f)
	require.NoError(t, err)
	defer m.Close()

	require.Equal(t, 11, m.Len())
	require.Equal(t, byte('h'), m.At(0))
	require.Equal(t, byte('d'), m.At(10))

	// UnsafeSlice - within bounds
	bz, err := m.UnsafeSlice(0, 5)
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), bz)

	// UnsafeSlice - within bounds
	bz, err = m.UnsafeSlice(6, 5)
	require.NoError(t, err)
	require.Equal(t, []byte("world"), bz)

	// UnsafeSlice - out of bounds
	_, err = m.UnsafeSlice(10, 5)
	require.Error(t, err)

	// UnsafeSliceVar - more than available
	n, bz, err := m.UnsafeSliceVar(0, 20)
	require.NoError(t, err)
	require.Equal(t, 11, n)
	require.Equal(t, []byte("hello world"), bz)

	// UnsafeSliceVar - less than available
	n, bz, err = m.UnsafeSliceVar(6, 3)
	require.NoError(t, err)
	require.Equal(t, 3, n)
	require.Equal(t, []byte("wor"), bz)

	// UnsafeSliceVar - exactly available
	n, bz, err = m.UnsafeSliceVar(0, 11)
	require.NoError(t, err)
	require.Equal(t, 11, n)
	require.Equal(t, []byte("hello world"), bz)

	// UnsafeSliceVar - no data
	_, _, err = m.UnsafeSliceVar(11, 5)
	require.Error(t, err)

	// UnsafeSliceVar - out of bounds
	_, _, err = m.UnsafeSliceVar(20, 5)
	require.Error(t, err)
}

func TestMmap_EmptyFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "empty")
	require.NoError(t, err)
	defer f.Close()

	m, err := NewMmap(f)
	require.NoError(t, err)
	require.Equal(t, 0, m.Len())
	require.NoError(t, m.Close())
}
