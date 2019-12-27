package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBitArrayClear(t *testing.T) {
	ba := NewBitArray(8)
	ba.Fill(7)
	require.True(t, ba.Get(7))
	ba.Clear(7)
	require.False(t, ba.Get(7))
}

func TestBitArrayFill(t *testing.T) {
	ba := NewBitArray(16)
	ba.Fill(8)
	ba.Fill(9)
	require.True(t, ba.Get(8))
	require.True(t, ba.Get(9))
}

func TestBitArrayHintAlloc(t *testing.T) {
	require.Len(t, NewBitArray(8), 1)
	require.Len(t, NewBitArray(9), 2)
}
