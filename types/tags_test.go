package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppendTags(t *testing.T) {
	a := SingleTag(MakeTag("a", []byte("1")))
	b := SingleTag(MakeTag("b", []byte("2")))
	c := AppendTags(a, b)
	require.Equal(t, c, Tags{MakeTag("a", []byte("1")), MakeTag("b", []byte("2"))})
}

func TestEmptyTags(t *testing.T) {
	a := EmptyTags()
	require.Equal(t, a, Tags{})
}

func TestSingleTag(t *testing.T) {
	a := MakeTag("a", []byte("1"))
	b := SingleTag(a)
	require.Equal(t, b, Tags{MakeTag("a", []byte("1"))})
}

func TestMakeTag(t *testing.T) {
	a := MakeTag("a", []byte("1"))
	require.Equal(t, a, Tag{[]byte("a"), []byte("1")})
}
