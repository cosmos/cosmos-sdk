package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendTags(t *testing.T) {
	a := NewTags("a", "1")
	b := NewTags("b", "2")
	c := a.AppendTags(b)
	require.Equal(t, c, Tags{MakeTag("a", "1"), MakeTag("b", "2")})
	require.Equal(t, c, Tags{MakeTag("a", "1")}.AppendTag("b", "2"))
}

func TestEmptyTags(t *testing.T) {
	a := EmptyTags()
	require.Equal(t, a, Tags{})
}

func TestNewTags(t *testing.T) {
	b := NewTags("a", "1")
	require.Equal(t, b, Tags{MakeTag("a", "1")})

	require.Panics(t, func() { NewTags("a", "1", "b") })
	require.Panics(t, func() { NewTags("a", 1) })
	require.Panics(t, func() { NewTags(1, 1) })
	require.Panics(t, func() { NewTags(true, false) })
}

func TestKVPairTags(t *testing.T) {
	a := NewTags("a", "1")
	require.Equal(t, a, Tags(a.ToKVPairs()))
}
